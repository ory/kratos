// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/ory/x/sqlcon"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/otelx"

	"github.com/samber/lo"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
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

func (s *Strategy) CompletedAuthenticationMethod(ctx context.Context, amr session.AuthenticationMethods) session.AuthenticationMethod {
	aal1Satisfied := lo.ContainsBy(amr, func(am session.AuthenticationMethod) bool {
		return am.Method != identity.CredentialsTypeCodeAuth && am.AAL == identity.AuthenticatorAssuranceLevel1
	})
	if aal1Satisfied {
		return session.AuthenticationMethod{
			Method: identity.CredentialsTypeCodeAuth,
			AAL:    identity.AuthenticatorAssuranceLevel2,
		}
	}
	return session.AuthenticationMethod{
		Method: identity.CredentialsTypeCodeAuth,
		AAL:    identity.AuthenticatorAssuranceLevel1,
	}
}

func (s *Strategy) HandleLoginError(r *http.Request, f *login.Flow, body *updateLoginFlowWithCodeMethod, err error) error {
	if errors.Is(err, flow.ErrCompletedByStrategy) {
		return err
	}

	if f != nil {
		email := ""
		if body != nil {
			email = body.Identifier
		}

		ds, err := s.deps.Config().DefaultIdentityTraitsSchemaURL(r.Context())
		if err != nil {
			return err
		}
		identifierLabel, err := login.GetIdentifierLabelFromSchema(r.Context(), ds.String())
		if err != nil {
			return err
		}
		f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
		f.UI.GetNodes().Upsert(
			node.NewInputField("identifier", email, node.DefaultGroup, node.InputAttributeTypeText, node.WithRequiredInputAttribute).
				WithMetaLabel(identifierLabel),
		)
	}

	return err
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, lf *login.Flow) error {
	return s.PopulateMethod(r, lf)
}

// findIdentityByIdentifier returns the identity and the code credential for the given identifier.
// If the identity does not have a code credential, it will attempt to find
// the identity through other credentials matching the identifier.
// the fallback mechanism is used for migration purposes of old accounts that do not have a code credential.
func (s *Strategy) findIdentityByIdentifier(ctx context.Context, identifier string) (_ *identity.Identity, isFallback bool, err error) {
	ctx, span := s.deps.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.code.strategy.findIdentityByIdentifier")
	defer otelx.End(span, &err)

	id, cred, err := s.deps.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, s.ID(), identifier)
	if errors.Is(err, sqlcon.ErrNoRows) {
		// this is a migration for old identities that do not have a code credential
		// we might be able to do a fallback login since we could not find a credential on this identifier
		// Case insensitive because we only care about emails.
		id, err := s.deps.PrivilegedIdentityPool().FindIdentityByCredentialIdentifier(ctx, identifier, false)
		if err != nil {
			return nil, false, errors.WithStack(schema.NewNoCodeAuthnCredentials())
		}

		// we don't know if the user has verified the code yet, so we just return the identity
		// and let the caller decide what to do with it
		return id, true, nil
	} else if err != nil {
		return nil, false, errors.WithStack(schema.NewNoCodeAuthnCredentials())
	}

	if len(cred.Identifiers) == 0 {
		return nil, false, errors.WithStack(schema.NewNoCodeAuthnCredentials())
	}

	// we don't need the code credential, we just need to know that it exists
	return id, false, nil
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, sess *session.Session) (_ *identity.Identity, err error) {
	ctx, span := s.deps.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.code.strategy.Login")
	defer otelx.End(span, &err)

	if err := flow.MethodEnabledAndAllowedFromRequest(r, f.GetFlowName(), s.ID().String(), s.deps); err != nil {
		return nil, err
	}

	var aal identity.AuthenticatorAssuranceLevel

	if s.deps.Config().SelfServiceCodeStrategy(ctx).PasswordlessEnabled {
		aal = identity.AuthenticatorAssuranceLevel1
	} else if s.deps.Config().SelfServiceCodeStrategy(ctx).MFAEnabled {
		aal = identity.AuthenticatorAssuranceLevel2
	}

	if err := login.CheckAAL(f, aal); err != nil {
		return nil, err
	}

	var p updateLoginFlowWithCodeMethod
	if err := s.dx.Decode(r, &p,
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPKeepRequestBody(true),
		decoderx.MustHTTPRawJSONSchemaCompiler(loginMethodSchema),
		decoderx.HTTPDecoderAllowedMethods("POST"),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return nil, s.HandleLoginError(r, f, &p, err)
	}

	if err := flow.EnsureCSRF(s.deps, r, f.Type, s.deps.Config().DisableAPIFlowEnforcement(ctx), s.deps.GenerateCSRFToken, p.CSRFToken); err != nil {
		return nil, s.HandleLoginError(r, f, &p, err)
	}

	// By Default the flow should be in the 'choose method' state.
	SetDefaultFlowState(f, p.Resend)

	switch f.GetState() {
	case flow.StateChooseMethod:
		if err := s.loginSendCode(ctx, w, r, f, &p, sess); err != nil {
			return nil, s.HandleLoginError(r, f, &p, err)
		}
		return nil, nil
	case flow.StateEmailSent:
		i, err := s.loginVerifyCode(ctx, r, f, &p)
		if err != nil {
			return nil, s.HandleLoginError(r, f, &p, err)
		}
		return i, nil
	case flow.StatePassedChallenge:
		return nil, s.HandleLoginError(r, f, &p, errors.WithStack(schema.NewNoLoginStrategyResponsible()))
	}

	return nil, s.HandleLoginError(r, f, &p, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unexpected flow state: %s", f.GetState())))
}

func (s *Strategy) loginSendCode(ctx context.Context, w http.ResponseWriter, r *http.Request, f *login.Flow, p *updateLoginFlowWithCodeMethod, sess *session.Session) (err error) {
	ctx, span := s.deps.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.code.strategy.loginSendCode")
	defer otelx.End(span, &err)

	if len(p.Identifier) == 0 {
		return errors.WithStack(schema.NewRequiredError("#/identifier", "identifier"))
	}

	p.Identifier = maybeNormalizeEmail(p.Identifier)

	var addresses []Address
	var i *identity.Identity
	if f.RequestedAAL > identity.AuthenticatorAssuranceLevel1 {
		address, found := lo.Find(sess.Identity.VerifiableAddresses, func(va identity.VerifiableAddress) bool {
			return va.Value == p.Identifier
		})
		if !found {
			return errors.WithStack(schema.NewUnknownAddressError())
		}
		i = sess.Identity
		addresses = []Address{{
			To:  address.Value,
			Via: address.Via,
		}}
	} else {
		// Step 1: Get the identity
		i, _, err = s.findIdentityByIdentifier(ctx, p.Identifier)
		if err != nil {
			return err
		}
		addresses = []Address{{
			To:  p.Identifier,
			Via: identity.CodeAddressType(identity.AddressTypeEmail),
		}}
	}

	// Step 2: Delete any previous login codes for this flow ID
	if err := s.deps.LoginCodePersister().DeleteLoginCodesOfFlow(ctx, f.GetID()); err != nil {
		return errors.WithStack(err)
	}

	// kratos only supports `email` identifiers at the moment with the code method
	// this is validated in the identity validation step above
	if err := s.deps.CodeSender().SendCode(ctx, f, i, addresses...); err != nil {
		return errors.WithStack(err)
	}

	// sets the flow state to code sent
	f.SetState(flow.NextState(f.GetState()))

	if err := s.NewCodeUINodes(r, f, &codeIdentifier{Identifier: p.Identifier}); err != nil {
		return err
	}

	f.Active = identity.CredentialsTypeCodeAuth
	if err = s.deps.LoginFlowPersister().UpdateLoginFlow(ctx, f); err != nil {
		return err
	}

	if x.IsJSONRequest(r) {
		s.deps.Writer().WriteCode(w, r, http.StatusBadRequest, f)
	} else {
		http.Redirect(w, r, f.AppendTo(s.deps.Config().SelfServiceFlowLoginUI(ctx)).String(), http.StatusSeeOther)
	}

	// we return an error to the flow handler so that it does not continue execution of the hooks.
	// we are not done with the login flow yet. The user needs to verify the code and then we need to persist the identity.
	return errors.WithStack(flow.ErrCompletedByStrategy)
}

// If identifier is an email, we lower case it because on mobile phones the first letter sometimes is capitalized.
func maybeNormalizeEmail(input string) string {
	if strings.Contains(input, "@") {
		return strings.ToLower(input)
	}
	return input
}

func (s *Strategy) loginVerifyCode(ctx context.Context, r *http.Request, f *login.Flow, p *updateLoginFlowWithCodeMethod) (_ *identity.Identity, err error) {
	ctx, span := s.deps.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.code.strategy.loginVerifyCode")
	defer otelx.End(span, &err)

	// we are in the second submission state of the flow
	// we need to check the code and update the identity
	if p.Code == "" {
		return nil, errors.WithStack(schema.NewRequiredError("#/code", "code"))
	}

	if len(p.Identifier) == 0 {
		return nil, errors.WithStack(schema.NewRequiredError("#/identifier", "identifier"))
	}

	p.Identifier = maybeNormalizeEmail(p.Identifier)

	isFallback := false
	var i *identity.Identity
	if f.RequestedAAL > identity.AuthenticatorAssuranceLevel1 {
		// Don't require the code credential if the user already has a session (e.g. this is an MFA flow)
		sess, err := s.deps.SessionManager().FetchFromRequest(ctx, r)
		if err != nil {
			return nil, err
		}
		i = sess.Identity
	} else {
		// Step 1: Get the identity
		i, isFallback, err = s.findIdentityByIdentifier(ctx, p.Identifier)
		if err != nil {
			return nil, err
		}
	}

	loginCode, err := s.deps.LoginCodePersister().UseLoginCode(ctx, f.ID, i.ID, p.Code)
	if err != nil {
		if errors.Is(err, ErrCodeNotFound) {
			return nil, schema.NewLoginCodeInvalid()
		}
		return nil, errors.WithStack(err)
	}

	i, err = s.deps.PrivilegedIdentityPool().GetIdentity(ctx, loginCode.IdentityID, identity.ExpandDefault)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// the code is correct, if the login happened through a different credential, we need to update the identity
	if isFallback {
		if err := i.SetCredentialsWithConfig(
			s.ID(),
			// p.Identifier was normalized prior.
			identity.Credentials{Type: s.ID(), Identifiers: []string{p.Identifier}},
			&identity.CredentialsCode{UsedAt: sql.NullTime{}},
		); err != nil {
			return nil, errors.WithStack(err)
		}

		if err := s.deps.PrivilegedIdentityPool().UpdateIdentity(ctx, i); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	// Step 2: The code was correct
	f.Active = identity.CredentialsTypeCodeAuth

	// since nothing has errored yet, we can assume that the code is correct
	// and we can update the login flow
	f.SetState(flow.NextState(f.GetState()))

	if err := s.deps.LoginFlowPersister().UpdateLoginFlow(ctx, f); err != nil {
		return nil, errors.WithStack(err)
	}

	for idx := range i.VerifiableAddresses {
		va := i.VerifiableAddresses[idx]
		if !va.Verified && loginCode.Address == va.Value {
			va.Verified = true
			va.Status = identity.VerifiableAddressStatusCompleted
			if err := s.deps.PrivilegedIdentityPool().UpdateVerifiableAddress(ctx, &va); err != nil {
				return nil, err
			}
			break
		}
	}

	return i, nil
}
