// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/herodot"
	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flowhelpers"
	"github.com/ory/kratos/selfservice/strategy/idfirst"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/text"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/stringsx"
)

var (
	_ login.AAL1FormHydrator = new(Strategy)
	_ login.Strategy         = new(Strategy)
)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
	s.setRoutes(r)
}

// Update Login Flow with OpenID Connect Method
//
// swagger:model updateLoginFlowWithOidcMethod
type UpdateLoginFlowWithOidcMethod struct {
	// The provider to register with
	//
	// required: true
	Provider string `json:"provider"`

	// The CSRF Token
	CSRFToken string `json:"csrf_token"`

	// Method to use
	//
	// This field must be set to `oidc` when using the oidc method.
	//
	// required: true
	Method string `json:"method"`

	// The identity traits. This is a placeholder for the registration flow.
	Traits json.RawMessage `json:"traits"`

	// UpstreamParameters are the parameters that are passed to the upstream identity provider.
	//
	// These parameters are optional and depend on what the upstream identity provider supports.
	// Supported parameters are:
	// - `login_hint` (string): The `login_hint` parameter suppresses the account chooser and either pre-fills the email box on the sign-in form, or selects the proper session.
	// - `hd` (string): The `hd` parameter limits the login/registration process to a Google Organization, e.g. `mycollege.edu`.
	// - `prompt` (string): The `prompt` specifies whether the Authorization Server prompts the End-User for reauthentication and consent, e.g. `select_account`.
	//
	// required: false
	UpstreamParameters json.RawMessage `json:"upstream_parameters"`

	// IDToken is an optional id token provided by an OIDC provider
	//
	// If submitted, it is verified using the OIDC provider's public key set and the claims are used to populate
	// the OIDC credentials of the identity.
	// If the OIDC provider does not store additional claims (such as name, etc.) in the IDToken itself, you can use
	// the `traits` field to populate the identity's traits. Note, that Apple only includes the users email in the IDToken.
	//
	// Supported providers are
	// - Apple
	// - Google
	// required: false
	IDToken string `json:"id_token,omitempty"`

	// IDTokenNonce is the nonce, used when generating the IDToken.
	// If the provider supports nonce validation, the nonce will be validated against this value and required.
	//
	// required: false
	IDTokenNonce string `json:"id_token_nonce,omitempty"`

	// Transient data to pass along to any webhooks
	//
	// required: false
	TransientPayload json.RawMessage `json:"transient_payload,omitempty" form:"transient_payload"`
}

func (s *Strategy) handleConflictingIdentity(ctx context.Context, w http.ResponseWriter, r *http.Request, loginFlow *login.Flow, token *identity.CredentialsOIDCEncryptedTokens, claims *Claims, provider Provider, container *AuthCodeContainer) (verdict ConflictingIdentityVerdict, id *identity.Identity, credentials *identity.Credentials, err error) {
	if s.conflictingIdentityPolicy == nil {
		return ConflictingIdentityVerdictReject, nil, nil, nil
	}

	// Find out if there is a conflicting identity
	newIdentity, va, err := s.newIdentityFromClaims(ctx, claims, provider, container, loginFlow.IdentitySchema)
	if err != nil {
		return ConflictingIdentityVerdictReject, nil, nil, nil
	}

	// Validate the identity itself
	// We ignore the error here because the claims may not fulfil the requirements
	// of the identity schema.
	//
	// However, this is not a problem because the identity will be merged with the existing
	// identity and the existing identity will be updated with the new credentials, but not any traits.
	//
	// We do need the validation step however, to "hydrate" the verifiable address of the user, which is then
	// used in subsequent calls to match the existing with the new identity.
	_ = s.d.IdentityValidator().Validate(ctx, newIdentity)

	for n := range newIdentity.VerifiableAddresses {
		verifiable := &newIdentity.VerifiableAddresses[n]
		for _, verified := range va {
			if verifiable.Via == verified.Via && verifiable.Value == verified.Value {
				verifiable.Status = identity.VerifiableAddressStatusCompleted
				verifiable.Verified = true
				t := sqlxx.NullTime(time.Now().UTC().Round(time.Second))
				verifiable.VerifiedAt = &t
			}
		}
	}

	creds, err := identity.NewOIDCLikeCredentials(token, s.ID(), provider.Config().ID, claims.Subject, provider.Config().OrganizationID)
	if err != nil {
		return ConflictingIdentityVerdictUnknown, nil, nil, err
	}

	newIdentity.SetCredentials(s.ID(), *creds)

	existingIdentity, _, _, err := s.d.IdentityManager().ConflictingIdentity(ctx, newIdentity)
	if err != nil {
		return ConflictingIdentityVerdictReject, nil, nil, nil
	}

	verdict = s.conflictingIdentityPolicy(ctx, existingIdentity, newIdentity, provider, claims)
	if verdict == ConflictingIdentityVerdictMerge {
		if err = existingIdentity.MergeOIDCCredentials(s.ID(), *creds); err != nil {
			return ConflictingIdentityVerdictUnknown, nil, nil, err
		}

		if err = s.d.PrivilegedIdentityPool().UpdateIdentity(ctx, existingIdentity); err != nil {
			return ConflictingIdentityVerdictUnknown, nil, nil, err
		}
	}

	return verdict, existingIdentity, creds, nil
}

func (s *Strategy) ProcessLogin(ctx context.Context, w http.ResponseWriter, r *http.Request, loginFlow *login.Flow, token *identity.CredentialsOIDCEncryptedTokens, claims *Claims, provider Provider, container *AuthCodeContainer) (_ *registration.Flow, err error) {
	ctx, span := s.d.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.oidc.Strategy.processLogin")
	defer otelx.End(span, &err)

	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, s.ID(), identity.OIDCUniqueID(provider.Config().ID, claims.Subject))
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
			var verdict ConflictingIdentityVerdict
			verdict, i, c, err = s.handleConflictingIdentity(ctx, w, r, loginFlow, token, claims, provider, container)
			switch verdict {
			case ConflictingIdentityVerdictMerge:
				// Do nothing
			case ConflictingIdentityVerdictReject:
				// If no account was found we're "manually" creating a new registration flow and redirecting the browser
				// to that endpoint.

				// That will execute the "pre registration" hook which allows to e.g. disallow this request. The registration
				// ui however will NOT be shown, instead the user is directly redirected to the auth path. That should then
				// do a silent re-request. While this might be a bit excessive from a network perspective it should usually
				// happen without any downsides to user experience as the flow has already been authorized and should
				// not need additional consent/login.

				// This is kinda hacky but the only way to ensure seamless login/registration flows when using OIDC.
				s.d.
					Logger().
					WithField("provider", provider.Config().ID).
					WithField("subject", claims.Subject).
					Debug("Received successful OpenID Connect callback but user is not registered. Re-initializing registration flow now.")

				// If return_to was set before, we need to preserve it.
				var opts []registration.FlowOption
				if len(loginFlow.ReturnTo) > 0 {
					opts = append(opts, registration.WithFlowReturnTo(loginFlow.ReturnTo))
				}

				if loginFlow.OAuth2LoginChallenge.String() != "" {
					opts = append(opts, registration.WithFlowOAuth2LoginChallenge(loginFlow.OAuth2LoginChallenge.String()))
				}

				registrationFlow, err := s.d.RegistrationHandler().NewRegistrationFlow(w, r, loginFlow.Type, opts...)
				if err != nil {
					return nil, s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, err)
				}

				err = s.d.SessionTokenExchangePersister().MoveToNewFlow(ctx, loginFlow.ID, registrationFlow.ID)
				if err != nil {
					return nil, s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, err)
				}

				registrationFlow.OrganizationID = loginFlow.OrganizationID
				registrationFlow.IDToken = loginFlow.IDToken
				registrationFlow.RawIDTokenNonce = loginFlow.RawIDTokenNonce
				registrationFlow.TransientPayload = loginFlow.TransientPayload
				registrationFlow.Active = s.ID()
				registrationFlow.IdentitySchema = loginFlow.IdentitySchema

				// We are converting the flow here, but want to retain the original request URL.
				registrationFlow.RequestURL = loginFlow.RequestURL

				if _, err := s.processRegistration(ctx, w, r, registrationFlow, token, claims, provider, container); err != nil {
					return registrationFlow, err
				}

				return nil, nil
			case ConflictingIdentityVerdictUnknown:
				fallthrough
			default:
				// This should never happen if err == nil, but just for safety:
				if err != nil {
					return nil, err
				}
				return nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("The OpenID Connect identity merge policy returned an unknown verdict without other error details, which prevents the sign up from completing. Please report this as a bug."))
			}

		} else {
			return nil, s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, err)
		}
	}

	var oidcCredentials identity.CredentialsOIDC
	if err := json.NewDecoder(bytes.NewBuffer(c.Config)).Decode(&oidcCredentials); err != nil {
		return nil, s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, x.WrapWithIdentityIDError(errors.WithStack(herodot.ErrInternalServerError.WithReason("The OpenID Connect credentials could not be decoded properly").WithDebug(err.Error())), i.ID))
	}

	sess := session.NewInactiveSession()
	sess.CompletedLoginForWithProvider(s.ID(), identity.AuthenticatorAssuranceLevel1, provider.Config().ID, provider.Config().OrganizationID)

	for _, c := range oidcCredentials.Providers {
		if c.Subject == claims.Subject && c.Provider == provider.Config().ID {
			if err = s.d.LoginHookExecutor().PostLoginHook(w, r, node.OpenIDConnectGroup, loginFlow, i, sess, provider.Config().ID); err != nil {
				return nil, x.WrapWithIdentityIDError(s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, err), i.ID)
			}
			return nil, nil
		}
	}

	return nil, s.HandleError(ctx, w, r, loginFlow, provider.Config().ID, nil, x.WrapWithIdentityIDError(errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to find matching OpenID Connect credentials.").WithDebugf(`Unable to find credentials that match the given provider "%s" and subject "%s".`, provider.Config().ID, claims.Subject)), i.ID))
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, _ *session.Session) (i *identity.Identity, err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.oidc.Strategy.Login")
	defer otelx.End(span, &err)

	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		span.SetAttributes(attribute.String("not_responsible_reason", "requested AAL is not AAL1"))
		return nil, err
	}

	var p UpdateLoginFlowWithOidcMethod
	if err := s.newLinkDecoder(ctx, &p, r, &f.IdentitySchema); err != nil {
		return nil, s.HandleError(ctx, w, r, f, "", nil, err)
	}

	f.IDToken = p.IDToken
	f.RawIDTokenNonce = p.IDTokenNonce
	f.TransientPayload = p.TransientPayload

	pid := p.Provider // this can come from both url query and post body
	if pid == "" {
		span.SetAttributes(attribute.String("not_responsible_reason", "provider ID missing"))
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if !strings.EqualFold(strings.ToLower(p.Method), s.SettingsStrategyID()) && p.Method != "" {
		// the user is sending a method that is not oidc, but the payload includes a provider
		s.d.Audit().
			WithRequest(r).
			WithField("provider", p.Provider).
			WithField("method", p.Method).
			Warn("The payload includes a `provider` field but is using a method other than `oidc`. Therefore, social sign in will not be executed.")
		span.SetAttributes(attribute.String("not_responsible_reason", "method is not oidc"))
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if err := flow.MethodEnabledAndAllowed(ctx, f.GetFlowName(), s.SettingsStrategyID(), s.SettingsStrategyID(), s.d); err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, s.handleMethodNotAllowedError(err))
	}

	provider, err := s.Provider(ctx, pid)
	if err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	}

	req, err := s.validateFlow(ctx, r, f.ID)
	if err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	}

	if authenticated, err := s.alreadyAuthenticated(ctx, w, r, req); err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	} else if authenticated {
		return i, nil
	}

	if p.IDToken != "" {
		claims, err := s.ProcessIDToken(r, provider, p.IDToken, p.IDTokenNonce)
		if err != nil {
			return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
		}
		_, err = s.ProcessLogin(ctx, w, r, f, nil, claims, provider, &AuthCodeContainer{
			FlowID: f.ID.String(),
			Traits: p.Traits,
		})
		if err != nil {
			return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
		}
		return nil, errors.WithStack(flow.ErrCompletedByStrategy)
	}

	state, pkce, err := s.GenerateState(ctx, provider, f.ID)
	if err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	}
	if err := s.d.ContinuityManager().Pause(ctx, w, r, sessionName,
		continuity.WithPayload(&AuthCodeContainer{
			State:            state,
			FlowID:           f.ID.String(),
			Traits:           p.Traits,
			TransientPayload: f.TransientPayload,
			IdentitySchema:   f.IdentitySchema,
		}),
		continuity.WithLifespan(time.Minute*30)); err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	}

	f.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(ctx, f); err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Could not update flow").WithWrap(err)))
	}

	var up map[string]string
	if err := json.NewDecoder(bytes.NewBuffer(p.UpstreamParameters)).Decode(&up); err != nil {
		return nil, err
	}

	codeURL, err := getAuthRedirectURL(ctx, provider, f, state, up, pkce)
	if err != nil {
		return nil, s.HandleError(ctx, w, r, f, pid, nil, err)
	}

	if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(codeURL))
	} else {
		http.Redirect(w, r, codeURL, http.StatusSeeOther)
	}

	return nil, errors.WithStack(flow.ErrCompletedByStrategy)
}

func (s *Strategy) PopulateLoginMethodFirstFactorRefresh(r *http.Request, lf *login.Flow, _ *session.Session) error {
	conf, err := s.Config(r.Context())
	if err != nil {
		return err
	}

	var providers []Configuration
	_, id, c := flowhelpers.GuessForcedLoginIdentifier(r, s.d, lf, s.ID())
	if id == nil || c == nil {
		providers = nil
	} else {
		var credentials identity.CredentialsOIDC
		if err := json.Unmarshal(c.Config, &credentials); err != nil {
			// failed to read OIDC credentials, don't add any providers
			providers = nil
		} else {
			// add only providers that can actually be used to log in as this identity
			providers = make([]Configuration, 0, len(conf.Providers))
			for i := range conf.Providers {
				for j := range credentials.Providers {
					if conf.Providers[i].ID == credentials.Providers[j].Provider {
						providers = append(providers, conf.Providers[i])
						break
					}
				}
			}
		}
	}

	lf.UI.SetCSRF(s.d.GenerateCSRFToken(r))
	AddProviders(lf.UI, providers, text.NewInfoLoginWith, s.ID())
	return nil
}

func (s *Strategy) PopulateLoginMethodFirstFactor(r *http.Request, f *login.Flow) error {
	return s.populateMethod(r, f, text.NewInfoLoginWith)
}

func (s *Strategy) removeProviders(conf *ConfigurationCollection, f *login.Flow) {
	for _, l := range conf.Providers {
		group := node.OpenIDConnectGroup
		if s.ID() == identity.CredentialsTypeSAML {
			group = node.SAMLGroup
		}

		if l.OrganizationID != "" {
			continue
		}

		f.GetUI().Nodes.RemoveMatching(&node.Node{
			Group: group,
			Type:  node.Input,
			Attributes: &node.InputAttributes{
				Name:       "provider",
				FieldValue: l.ID,
			},
		})
	}
}

func (s *Strategy) PopulateLoginMethodIdentifierFirstCredentials(r *http.Request, f *login.Flow, mods ...login.FormHydratorModifier) (err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.oidc.Strategy.PopulateLoginMethodIdentifierFirstCredentials")
	defer otelx.End(span, &err)

	conf, err := s.Config(ctx)
	if err != nil {
		return err
	}

	o := login.NewFormHydratorOptions(mods)

	var linked []Provider
	if o.IdentityHint != nil {
		var err error
		// If we have an identity hint we check if the identity has any providers configured.
		if linked, err = s.linkedProviders(conf, o.IdentityHint); err != nil {
			return err
		}
	}

	if len(linked) == 0 {
		// If we found no credentials:
		if s.d.Config().SecurityAccountEnumerationMitigate(ctx) {
			// We found no credentials but do not want to leak that we know that. So we return early and do not
			// modify the initial provider list.
			return nil
		}

		if o.IdentityHint != nil {
			// We found no credentials. We remove all the providers and tell the strategy that we found nothing.
			// We only execute this, if the identity hint is set, otherwise we do not know if the user has any credentials and we likely stay on the `provide_credentials` screen.
			// The OIDC method is special in that regard, as it's the only method showing buttons on that screen.
			s.removeProviders(conf, f)
		}
		return idfirst.ErrNoCredentialsFound
	}

	if !s.d.Config().SecurityAccountEnumerationMitigate(ctx) {
		// Account enumeration is disabled, so we show all providers that are linked to the identity.
		// User is found and enumeration mitigation is disabled. Filter the list!
		s.removeProviders(conf, f)

		for _, l := range linked {
			lc := l.Config()

			// Organizations are handled differently.
			if lc.OrganizationID != "" {
				continue
			}
			AddProvider(f.UI, lc.ID, text.NewInfoLoginWith(stringsx.Coalesce(lc.Label, lc.ID), lc.ID), s.ID())
		}
	}

	return nil
}

func (s *Strategy) PopulateLoginMethodIdentifierFirstIdentification(r *http.Request, f *login.Flow) error {
	return s.populateMethod(r, f, text.NewInfoLoginWith)
}
