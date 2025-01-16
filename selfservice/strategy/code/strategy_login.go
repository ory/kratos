// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"cmp"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/kratos/driver/config"

	"github.com/ory/kratos/selfservice/strategy/idfirst"
	"github.com/ory/kratos/text"

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

var (
	_ login.FormHydrator = new(Strategy)
	_ login.Strategy     = new(Strategy)
)

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

	// Address is the address to send the code to, in case that there are multiple addresses. This field
	// is only used in two-factor flows and is ineffective for passwordless flows.
	Address string `json:"address" form:"address"`

	// Resend is set when the user wants to resend the code
	// required: false
	Resend string `json:"resend" form:"resend"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}

func (s *Strategy) RegisterLoginRoutes(*x.RouterPublic) {}

func (s *Strategy) CompletedAuthenticationMethod(ctx context.Context) session.AuthenticationMethod {
	aal := identity.AuthenticatorAssuranceLevel1
	if s.deps.Config().SelfServiceCodeStrategy(ctx).MFAEnabled {
		aal = identity.AuthenticatorAssuranceLevel2
	}

	return session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    aal,
	}
}

func (s *Strategy) HandleLoginError(r *http.Request, f *login.Flow, body *updateLoginFlowWithCodeMethod, err error) error {
	if errors.Is(err, flow.ErrCompletedByStrategy) {
		return err
	}

	if f != nil {
		identifier := ""
		if body != nil {
			identifier = cmp.Or(body.Address, body.Identifier)
		}

		f.UI.SetCSRF(s.deps.GenerateCSRFToken(r))
		identifierNode := node.NewInputField("identifier", identifier, node.DefaultGroup, node.InputAttributeTypeHidden)

		identifierNode.Attributes.SetValue(identifier)
		f.UI.GetNodes().Upsert(identifierNode)
	}

	return err
}

// findIdentityByIdentifier returns the identity and the code credential for the given identifier.
// If the identity does not have a code credential, it will attempt to find
// the identity through other credentials matching the identifier.
// the fallback mechanism is used for migration purposes of old accounts that do not have a code credential.
func (s *Strategy) findIdentityByIdentifier(ctx context.Context, identifier string) (id *identity.Identity, cred *identity.Credentials, isFallback bool, err error) {
	ctx, span := s.deps.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.code.Strategy.findIdentityByIdentifier")
	defer otelx.End(span, &err)

	id, cred, err = s.deps.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, s.ID(), identifier)
	if errors.Is(err, sqlcon.ErrNoRows) {
		// this is a migration for old identities that do not have a code credential
		// we might be able to do a fallback login since we could not find a credential on this identifier
		// Case insensitive because we only care about emails.
		id, err := s.deps.PrivilegedIdentityPool().FindIdentityByCredentialIdentifier(ctx, identifier, false)
		if err != nil {
			return nil, nil, false, errors.WithStack(schema.NewNoCodeAuthnCredentials())
		}

		// we don't know if the user has verified the code yet, so we just return the identity
		// and let the caller decide what to do with it
		return id, nil, true, nil
	} else if err != nil {
		return nil, nil, false, errors.WithStack(schema.NewNoCodeAuthnCredentials())
	}

	if len(cred.Identifiers) == 0 {
		return nil, nil, false, errors.WithStack(schema.NewNoCodeAuthnCredentials())
	}

	// we don't need the code credential, we just need to know that it exists
	return id, cred, false, nil
}

type decodedMethod struct {
	Method  string `json:"method" form:"method"`
	Address string `json:"address" form:"address"`
}

func (s *Strategy) methodEnabledAndAllowedFromRequest(r *http.Request, f *login.Flow) (*decodedMethod, error) {
	var method decodedMethod

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(loginMethodSchema)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := decoderx.NewHTTP().Decode(r, &method, compiler,
		decoderx.HTTPKeepRequestBody(true),
		decoderx.HTTPDecoderAllowedMethods("POST", "PUT", "PATCH", "GET"),
		decoderx.HTTPDecoderSetValidatePayloads(false),
		decoderx.HTTPDecoderJSONFollowsFormFormat()); err != nil {
		return &method, errors.WithStack(err)
	}

	if err := flow.MethodEnabledAndAllowed(r.Context(), f.GetFlowName(), s.ID().String(), method.Method, s.deps); err != nil {
		return &method, err
	}

	return &method, nil
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, sess *session.Session) (_ *identity.Identity, err error) {
	ctx, span := s.deps.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.code.Strategy.Login")
	defer otelx.End(span, &err)

	if s.deps.Config().SelfServiceCodeStrategy(ctx).PasswordlessEnabled {
		if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
			return nil, err
		}
	} else if s.deps.Config().SelfServiceCodeStrategy(ctx).MFAEnabled {
		if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel2); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if p, err := s.methodEnabledAndAllowedFromRequest(r, f); errors.Is(err, flow.ErrStrategyNotResponsible) {
		if !s.deps.Config().SelfServiceCodeStrategy(ctx).MFAEnabled {
			span.SetAttributes(attribute.String("not_responsible_reason", "MFA is not enabled"))
			return nil, err
		}

		if p == nil || len(p.Address) == 0 {
			span.SetAttributes(attribute.String("not_responsible_reason", "method not set or address not set"))
			return nil, err
		}

		// In this special case we only expect `address` to be set.
	} else if err != nil {
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

	f.TransientPayload = p.TransientPayload

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
		i, err := s.loginVerifyCode(ctx, f, &p, sess)
		if err != nil {
			return nil, s.HandleLoginError(r, f, &p, err)
		}
		return i, nil
	case flow.StatePassedChallenge:
		return nil, s.HandleLoginError(r, f, &p, errors.WithStack(schema.NewNoLoginStrategyResponsible()))
	}

	return nil, s.HandleLoginError(r, f, &p, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unexpected flow state: %s", f.GetState())))
}

func (s *Strategy) findIdentifierInVerifiableAddress(i *identity.Identity, identifier string) (*Address, error) {
	verifiableAddress, found := lo.Find(i.VerifiableAddresses, func(va identity.VerifiableAddress) bool {
		return va.Value == identifier
	})
	if !found {
		return nil, errors.WithStack(schema.NewUnknownAddressError())
	}

	// This should be fine for legacy cases because we use `UpgradeCredentials` to normalize all address types prior
	// to calling this method.
	parsed, err := identity.NewCodeChannel(verifiableAddress.Via)
	if err != nil {
		return nil, err
	}

	return &Address{
		To:  verifiableAddress.Value,
		Via: parsed,
	}, nil
}

func (s *Strategy) findIdentityForIdentifier(ctx context.Context, identifier string, requestedAAL identity.AuthenticatorAssuranceLevel, session *session.Session) (_ *identity.Identity, _ []Address, err error) {
	ctx, span := s.deps.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.code.Strategy.findIdentityForIdentifier")
	span.SetAttributes(
		attribute.String("flow.requested_aal", string(requestedAAL)),
	)

	defer otelx.End(span, &err)

	if len(identifier) == 0 {
		return nil, nil, errors.WithStack(schema.NewRequiredError("#/identifier", "identifier"))
	}

	identifier = x.GracefulNormalization(identifier)

	var addresses []Address

	// Step 1: Get the identity
	i, cred, isFallback, err := s.findIdentityByIdentifier(ctx, identifier)
	if err != nil {
		if requestedAAL == identity.AuthenticatorAssuranceLevel2 {
			// When using two-factor auth, the identity used to not have any code credential associated. Therefore,
			// we need to gracefully handle this flow.
			//
			// TODO this section should be removed at some point when we are sure that all identities have a code credential.
			if codeCred := new(schema.ValidationError); errors.As(err, &codeCred) && codeCred.ValidationError.Message == "account does not exist or has not setup up sign in with code" {
				fallbackAllowed := s.deps.Config().SelfServiceCodeMethodMissingCredentialFallbackEnabled(ctx)
				span.SetAttributes(
					attribute.Bool(config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, fallbackAllowed),
				)

				if !fallbackAllowed {
					s.deps.Logger().Warn("The identity does not have a code credential but the fallback mechanism is disabled. Login failed.")
					return nil, nil, errors.WithStack(schema.NewNoCodeAuthnCredentials())
				}

				address, err := s.findIdentifierInVerifiableAddress(session.Identity, identifier)
				if err != nil {
					return nil, nil, err
				}

				// We only end up here if the identity's identity schema does not have the `code` identifier extension defined.
				// We know that this is the case for a couple of projects who use 2FA with the code credential.
				//
				// In those scenarios, the identity has no code credential, and the code credential will also not be created by
				// the identity schema.
				//
				// To avoid future regressions, we will not perform an update on the identity here. Effectively, whenever
				// the identity would be updated again (and the identity schema + extensions parsed), it would be likely
				// that the code credentials are overwritten.
				//
				// So we accept that the identity in this case will simply not have code credentials, and we will rely on the
				// fallback mechanism to authenticate the user.
				return session.Identity, []Address{*address}, nil
			} else if err != nil {
				return nil, nil, err
			}
		}
		return nil, nil, err
	} else if isFallback {
		fallbackAllowed := s.deps.Config().SelfServiceCodeMethodMissingCredentialFallbackEnabled(ctx)
		span.SetAttributes(
			attribute.String("identity.id", i.ID.String()),
			attribute.String("network.id", i.NID.String()),
			attribute.Bool(config.ViperKeyCodeConfigMissingCredentialFallbackEnabled, fallbackAllowed),
		)

		if !fallbackAllowed {
			s.deps.Logger().Warn("The identity does not have a code credential but the fallback mechanism is disabled. Login failed.")
			return nil, nil, errors.WithStack(schema.NewNoCodeAuthnCredentials())
		}

		// We don't have a code credential, but we can still login the user if they have a verified address.
		// This is a migration path for old accounts that do not have a code credential.
		addresses = []Address{{
			To:  identifier,
			Via: identity.CodeChannelEmail,
		}}

		// We only end up here if the identity's identity schema does not have the `code` identifier extension defined.
		// We know that this is the case for a couple of projects who use 2FA with the code credential.
		//
		// In those scenarios, the identity has no code credential, and the code credential will also not be created by
		// the identity schema.
		//
		// To avoid future regressions, we will not perform an update on the identity here. Effectively, whenever
		// the identity would be updated again (and the identity schema + extensions parsed), it would be likely
		// that the code credentials are overwritten.
		//
		// So we accept that the identity in this case will simply not have code credentials, and we will rely on the
		// fallback mechanism to authenticate the user.
	} else {
		span.SetAttributes(
			attribute.String("identity.id", i.ID.String()),
			attribute.String("network.id", i.NID.String()),
		)

		var conf identity.CredentialsCode
		if err := json.Unmarshal(cred.Config, &conf); err != nil {
			return nil, nil, errors.WithStack(err)
		}

		for _, address := range conf.Addresses {
			addresses = append(addresses, Address{
				To:  address.Address,
				Via: address.Channel,
			})
		}
	}

	return i, addresses, nil
}

func (s *Strategy) loginSendCode(ctx context.Context, w http.ResponseWriter, r *http.Request, f *login.Flow, p *updateLoginFlowWithCodeMethod, sess *session.Session) (err error) {
	ctx, span := s.deps.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.code.Strategy.loginSendCode")
	defer otelx.End(span, &err)

	p.Identifier = maybeNormalizeEmail(
		cmp.Or(p.Identifier, p.Address),
	)

	i, addresses, err := s.findIdentityForIdentifier(ctx, p.Identifier, f.RequestedAAL, sess)
	if err != nil {
		return err
	}

	if address, found := lo.Find(addresses, func(item Address) bool {
		return item.To == x.GracefulNormalization(p.Identifier)
	}); found {
		addresses = []Address{address}
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

func (s *Strategy) loginVerifyCode(ctx context.Context, f *login.Flow, p *updateLoginFlowWithCodeMethod, sess *session.Session) (_ *identity.Identity, err error) {
	ctx, span := s.deps.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.code.Strategy.loginVerifyCode")
	defer otelx.End(span, &err)

	// we are in the second submission state of the flow
	// we need to check the code and update the identity
	if p.Code == "" {
		return nil, errors.WithStack(schema.NewRequiredError("#/code", "code"))
	}

	p.Identifier = maybeNormalizeEmail(
		cmp.Or(
			p.Address,
			p.Identifier, // Older versions of Kratos required us to send the identifier here.
		),
	)

	var i *identity.Identity
	if f.RequestedAAL == identity.AuthenticatorAssuranceLevel2 {
		i = sess.Identity
	} else {
		i, _, err = s.findIdentityForIdentifier(ctx, p.Identifier, f.RequestedAAL, sess)
	}
	if err != nil {
		return nil, err
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

	// Step 2: The code was correct
	f.Active = identity.CredentialsTypeCodeAuth

	// since nothing has errored yet, we can assume that the code is correct
	// and we can update the login flow
	f.SetState(flow.NextState(f.GetState()))

	if err := s.deps.LoginFlowPersister().UpdateLoginFlow(ctx, f); err != nil {
		return nil, errors.WithStack(err)
	}

	// Step 3: Verify the address
	if err := s.verifyAddress(ctx, i, Address{
		To:  loginCode.Address,
		Via: loginCode.AddressType,
	}); err != nil {
		return nil, err
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

func (s *Strategy) verifyAddress(ctx context.Context, i *identity.Identity, verified Address) error {
	for idx := range i.VerifiableAddresses {
		va := i.VerifiableAddresses[idx]
		if va.Verified {
			continue
		}

		if verified.To != va.Value || string(verified.Via) != va.Via {
			continue
		}

		va.Verified = true
		va.Status = identity.VerifiableAddressStatusCompleted
		if err := s.deps.PrivilegedIdentityPool().UpdateVerifiableAddress(ctx, &va); errors.Is(err, sqlcon.ErrNoRows) {
			// This happens when the verified address does not yet exist, for example during registration. In this case we just skip.
			continue
		} else if err != nil {
			return err
		}
		break
	}

	return nil
}

func (s *Strategy) PopulateLoginMethodFirstFactorRefresh(r *http.Request, f *login.Flow) error {
	return s.PopulateMethod(r, f)
}

func (s *Strategy) PopulateLoginMethodFirstFactor(r *http.Request, f *login.Flow) error {
	return s.PopulateMethod(r, f)
}

func (s *Strategy) PopulateLoginMethodSecondFactor(r *http.Request, f *login.Flow) error {
	return s.PopulateMethod(r, f)
}

func (s *Strategy) PopulateLoginMethodSecondFactorRefresh(r *http.Request, f *login.Flow) error {
	return s.PopulateMethod(r, f)
}

func (s *Strategy) PopulateLoginMethodIdentifierFirstCredentials(r *http.Request, f *login.Flow, opts ...login.FormHydratorModifier) error {
	if !s.deps.Config().SelfServiceCodeStrategy(r.Context()).PasswordlessEnabled {
		// We only return this if passwordless is disabled, because if it is enabled we can always sign in using this method.
		return errors.WithStack(idfirst.ErrNoCredentialsFound)
	}
	o := login.NewFormHydratorOptions(opts)

	// If the identity hint is nil and account enumeration mitigation is disabled, we return an error.
	if o.IdentityHint == nil && !s.deps.Config().SecurityAccountEnumerationMitigate(r.Context()) {
		return errors.WithStack(idfirst.ErrNoCredentialsFound)
	}

	f.GetUI().Nodes.Append(
		node.NewInputField("method", s.ID(), node.CodeGroup, node.InputAttributeTypeSubmit).WithMetaLabel(text.NewInfoSelfServiceLoginCode()),
	)
	return nil
}

func (s *Strategy) PopulateLoginMethodIdentifierFirstIdentification(r *http.Request, f *login.Flow) error {
	return nil
}
