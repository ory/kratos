// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"net/http"

	"github.com/ory/herodot"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
)

var _ login.Strategy = new(Strategy)

// Update Login flow using the code method
//
// swagger:model updateLoginFlowWithCodeMethod
type updateLoginFlowWithCodeMethod struct {
	// Method should be set to "code" when logging in using the code strategy.
	//
	// required: true
	Method string `json:"method" form:"method"`

	// CSRFToken is the anti-CSRF token
	//
	// required: true
	CSRFToken string `json:"csrf_token" form:"csrf_token"`

	// Code is the 6 digits code sent to the user
	//
	// required: false
	Code string `json:"code" form:"code"`

	// Identifier is the code identifier
	// The identifier requires that the user has already completed the registration or settings with code flow.
	// required: false
	Identifier string `json:"identifier" form:"identifier"`

	// Resend is set when the user wants to resend the code
	// required: false
	Resend string `json:"resend" form:"resend"`
}

func (s *Strategy) RegisterLoginRoutes(*x.RouterPublic) {}

func (s *Strategy) CompletedAuthenticationMethod(ctx context.Context) session.AuthenticationMethod {
	return session.AuthenticationMethod{
		Method: identity.CredentialsTypeCodeAuth,
		AAL:    identity.AuthenticatorAssuranceLevel1,
	}
}

func (s *Strategy) HandleLoginError(w http.ResponseWriter, r *http.Request, flow *login.Flow, body *updateLoginFlowWithCodeMethod, err error) error {
	if flow != nil {
		email := ""
		if body != nil {
			email = body.Identifier
		}

		flow.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
		flow.UI.GetNodes().Upsert(
			node.NewInputField("identifier", email, node.DefaultGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).
				WithMetaLabel(text.NewInfoNodeLabelID()),
		)
	}

	return err
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, lf *login.Flow) error {
	return s.PopulateMethod(r, lf)
}

func (s *Strategy) getIdentity(ctx context.Context, identifier string) (*identity.Identity, *identity.Credentials, error) {
	i, cred, err := s.deps.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, s.ID(), identifier)
	if err != nil {
		return nil, nil, errors.WithStack(schema.NewNoCodeAuthnCredentials())
	}

	if len(cred.Identifiers) == 0 {
		return nil, nil, errors.WithStack(schema.NewNoCodeAuthnCredentials())
	} else if cred.IdentifierAddressType == "" {
		return nil, nil, errors.WithStack(schema.NewNoCodeAuthnCredentials())
	}

	return i, cred, nil
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, identityID uuid.UUID) (i *identity.Identity, err error) {
	s.deps.Audit().
		WithRequest(r).
		WithField("identity_id", identityID).
		WithField("login_flow_id", f.ID).
		Info("Login with the code strategy started.")

	if err := flow.MethodEnabledAndAllowedFromRequest(r, f.GetFlowName(), s.ID().String(), s.deps); err != nil {
		return nil, err
	}

	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, err
	}

	var p updateLoginFlowWithCodeMethod
	if err := s.dx.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPKeepRequestBody(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginMethodSchema),
		decoderx.HTTPDecoderAllowedMethods("POST"),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.HandleLoginError(w, r, f, &p, err)
	}

	if err := flow.EnsureCSRF(s.deps, r, f.Type, s.deps.Config().DisableAPIFlowEnforcement(r.Context()), s.deps.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.HandleLoginError(w, r, f, &p, err)
	}

	codeManager := NewCodeStateManager(f, s, &p)

	codeManager.SetCreateCodeHandler(func(ctx context.Context, f *login.Flow, strategy *Strategy, p *updateLoginFlowWithCodeMethod) error {
		strategy.deps.Audit().
			WithSensitiveField("identifier", p.Identifier).
			Info("Creating login code state.")

		// Step 1: Get the identity
		i, cred, err := strategy.getIdentity(ctx, p.Identifier)
		if err != nil {
			return err
		}

		// Step 2: Delete any previous login codes for this flow ID
		if err := strategy.deps.LoginCodePersister().DeleteLoginCodesOfFlow(ctx, f.GetID()); err != nil {
			return errors.WithStack(err)
		}

		addresses := []Address{
			{
				To:  p.Identifier,
				Via: identity.CodeAddressType(cred.IdentifierAddressType),
			},
		}

		// kratos only supports `email` identifiers at the moment with the code method
		// this is validated in the identity validation step above
		if err := strategy.deps.CodeSender().SendCode(ctx, f, i, addresses...); err != nil {
			return errors.WithStack(err)
		}

		// sets the flow state to code sent
		s.NextFlowState(f)

		if err := s.NewCodeUINodes(r, f, &codeIdentifier{Identifier: p.Identifier}); err != nil {
			return err
		}

		f.Active = identity.CredentialsTypeCodeAuth
		if err = strategy.deps.LoginFlowPersister().UpdateLoginFlow(ctx, f); err != nil {
			return err
		}

		if x.IsJSONRequest(r) {
			strategy.deps.Writer().WriteCode(w, r, http.StatusBadRequest, f)
		} else {
			http.Redirect(w, r, f.AppendTo(strategy.deps.Config().SelfServiceFlowLoginUI(ctx)).String(), http.StatusSeeOther)
		}

		// we return an error to the flow handler so that it does not continue execution of the hooks.
		// we are not done with the login flow yet. The user needs to verify the code and then we need to persist the identity.
		return errors.WithStack(flow.ErrCompletedByStrategy)
	})

	codeManager.SetCodeVerifyHandler(func(ctx context.Context, f *login.Flow, strategy *Strategy, p *updateLoginFlowWithCodeMethod) error {
		strategy.deps.Audit().
			WithSensitiveField("code", p.Code).
			WithSensitiveField("identifier", p.Identifier).
			Debug("Verifying login code")

		// Step 1: Get the identity
		i, _, err = strategy.getIdentity(ctx, p.Identifier)
		if err != nil {
			return err
		}

		loginCode, err := strategy.deps.LoginCodePersister().UseLoginCode(ctx, f.ID, i.ID, p.Code)
		if err != nil {
			if errors.Is(err, ErrCodeNotFound) {
				return schema.NewLoginCodeInvalid()
			}
			return errors.WithStack(err)
		}

		i, err = strategy.deps.PrivilegedIdentityPool().GetIdentity(ctx, loginCode.IdentityID, identity.ExpandDefault)
		if err != nil {
			return errors.WithStack(err)
		}

		// Step 2: The code was correct
		f.Active = identity.CredentialsTypeCodeAuth

		// since nothing has errored yet, we can assume that the code is correct
		// and we can update the login flow
		strategy.NextFlowState(f)

		if err := strategy.deps.LoginFlowPersister().UpdateLoginFlow(ctx, f); err != nil {
			return errors.WithStack(err)
		}

		for idx := range i.VerifiableAddresses {
			va := i.VerifiableAddresses[idx]
			if !va.Verified && loginCode.Address == va.Value {
				va.Verified = true
				va.Status = identity.VerifiableAddressStatusCompleted
				if err := strategy.deps.PrivilegedIdentityPool().UpdateVerifiableAddress(r.Context(), &va); err != nil {
					return err
				}
				break
			}
		}

		return nil
	})

	codeManager.SetCodeDoneHandler(func(ctx context.Context, f *login.Flow, strategy *Strategy, p *updateLoginFlowWithCodeMethod) error {
		strategy.deps.Audit().
			WithSensitiveField("identifier", p.Identifier).
			Debug("The login flow has already been completed, but is being re-requested.")
		return errors.WithStack(schema.NewNoLoginStrategyResponsible())
	})

	if err := codeManager.Run(r.Context()); err != nil {
		if errors.Is(err, flow.ErrCompletedByStrategy) {
			return nil, err
		}
		// the error is already handled by the registered code states
		return i, s.HandleLoginError(w, r, f, &p, err)
	}

	// a precaution in case the code manager did not set the identity
	if i == nil {
		s.deps.Audit().
			WithSensitiveField("identifier", p.Identifier).
			WithRequest(r).
			WithField("login_flow", f).
			Error("The code manager did not set the identity.")
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("the login flow did not complete successfully"))
	}

	return i, nil
}
