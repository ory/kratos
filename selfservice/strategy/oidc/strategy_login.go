// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/kratos/session"

	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/otelx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/selfservice/flow/registration"

	"github.com/ory/kratos/text"

	"github.com/ory/kratos/continuity"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

var _ login.Strategy = new(Strategy)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
	s.setRoutes(r)
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, requestedAAL identity.AuthenticatorAssuranceLevel, l *login.Flow) error {
	// This strategy can only solve AAL1
	if requestedAAL > identity.AuthenticatorAssuranceLevel1 {
		return nil
	}

	return s.populateMethod(r, l, text.NewInfoLoginWith)
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

func (s *Strategy) processLogin(w http.ResponseWriter, r *http.Request, loginFlow *login.Flow, token *identity.CredentialsOIDCEncryptedTokens, claims *Claims, provider Provider, container *AuthCodeContainer) (*registration.Flow, error) {
	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), identity.CredentialsTypeOIDC, identity.OIDCUniqueID(provider.Config().ID, claims.Subject))
	if err != nil {
		if errors.Is(err, sqlcon.ErrNoRows) {
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
				return nil, s.handleError(w, r, loginFlow, provider.Config().ID, nil, err)
			}

			err = s.d.SessionTokenExchangePersister().MoveToNewFlow(r.Context(), loginFlow.ID, registrationFlow.ID)
			if err != nil {
				return nil, s.handleError(w, r, loginFlow, provider.Config().ID, nil, err)
			}

			registrationFlow.OrganizationID = loginFlow.OrganizationID
			registrationFlow.IDToken = loginFlow.IDToken
			registrationFlow.RawIDTokenNonce = loginFlow.RawIDTokenNonce
			registrationFlow.RequestURL, err = x.TakeOverReturnToParameter(loginFlow.RequestURL, registrationFlow.RequestURL)
			registrationFlow.TransientPayload = loginFlow.TransientPayload
			if err != nil {
				return nil, s.handleError(w, r, loginFlow, provider.Config().ID, nil, err)
			}

			if _, err := s.processRegistration(w, r, registrationFlow, token, claims, provider, container, loginFlow.IDToken); err != nil {
				return registrationFlow, err
			}

			return nil, nil
		}

		return nil, s.handleError(w, r, loginFlow, provider.Config().ID, nil, err)
	}

	var oidcCredentials identity.CredentialsOIDC
	if err := json.NewDecoder(bytes.NewBuffer(c.Config)).Decode(&oidcCredentials); err != nil {
		return nil, s.handleError(w, r, loginFlow, provider.Config().ID, nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error())))
	}

	sess := session.NewInactiveSession()
	sess.CompletedLoginForWithProvider(s.ID(), identity.AuthenticatorAssuranceLevel1, provider.Config().ID,
		httprouter.ParamsFromContext(r.Context()).ByName("organization"))
	for _, c := range oidcCredentials.Providers {
		if c.Subject == claims.Subject && c.Provider == provider.Config().ID {
			if err = s.d.LoginHookExecutor().PostLoginHook(w, r, node.OpenIDConnectGroup, loginFlow, i, sess, provider.Config().ID); err != nil {
				return nil, s.handleError(w, r, loginFlow, provider.Config().ID, nil, err)
			}
			return nil, nil
		}
	}

	return nil, s.handleError(w, r, loginFlow, provider.Config().ID, nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to find matching OpenID Connect Credentials.").WithDebugf(`Unable to find credentials that match the given provider "%s" and subject "%s".`, provider.Config().ID, claims.Subject)))
}

func (s *Strategy) Login(w http.ResponseWriter, r *http.Request, f *login.Flow, _ *session.Session) (i *identity.Identity, err error) {
	ctx, span := s.d.Tracer(r.Context()).Tracer().Start(r.Context(), "selfservice.strategy.oidc.strategy.Login")
	defer otelx.End(span, &err)

	if err := login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1); err != nil {
		return nil, err
	}

	var p UpdateLoginFlowWithOidcMethod
	if err := s.newLinkDecoder(&p, r); err != nil {
		return nil, s.handleError(w, r, f, "", nil, err)
	}

	f.IDToken = p.IDToken
	f.RawIDTokenNonce = p.IDTokenNonce
	f.TransientPayload = p.TransientPayload

	pid := p.Provider // this can come from both url query and post body
	if pid == "" {
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if !strings.EqualFold(strings.ToLower(p.Method), s.SettingsStrategyID()) && p.Method != "" {
		// the user is sending a method that is not oidc, but the payload includes a provider
		s.d.Audit().
			WithRequest(r).
			WithField("provider", p.Provider).
			WithField("method", p.Method).
			Warn("The payload includes a `provider` field but is using a method other than `oidc`. Therefore, social sign in will not be executed.")
		return nil, errors.WithStack(flow.ErrStrategyNotResponsible)
	}

	if err := flow.MethodEnabledAndAllowed(ctx, f.GetFlowName(), s.SettingsStrategyID(), s.SettingsStrategyID(), s.d); err != nil {
		return nil, s.handleError(w, r, f, pid, nil, err)
	}

	provider, err := s.provider(ctx, r, pid)
	if err != nil {
		return nil, s.handleError(w, r, f, pid, nil, err)
	}

	req, err := s.validateFlow(ctx, r, f.ID)
	if err != nil {
		return nil, s.handleError(w, r, f, pid, nil, err)
	}

	if authenticated, err := s.alreadyAuthenticated(w, r, req); err != nil {
		return nil, s.handleError(w, r, f, pid, nil, err)
	} else if authenticated {
		return i, nil
	}

	if p.IDToken != "" {
		claims, err := s.processIDToken(w, r, provider, p.IDToken, p.IDTokenNonce)
		if err != nil {
			return nil, s.handleError(w, r, f, pid, nil, err)
		}
		_, err = s.processLogin(w, r, f, nil, claims, provider, &AuthCodeContainer{
			FlowID: f.ID.String(),
			Traits: p.Traits,
		})
		if err != nil {
			return nil, s.handleError(w, r, f, pid, nil, err)
		}
		return nil, errors.WithStack(flow.ErrCompletedByStrategy)
	}

	state := generateState(f.ID.String())
	if code, hasCode, _ := s.d.SessionTokenExchangePersister().CodeForFlow(ctx, f.ID); hasCode {
		state.setCode(code.InitCode)
	}
	if err := s.d.ContinuityManager().Pause(ctx, w, r, sessionName,
		continuity.WithPayload(&AuthCodeContainer{
			State:            state.String(),
			FlowID:           f.ID.String(),
			Traits:           p.Traits,
			TransientPayload: f.TransientPayload,
		}),
		continuity.WithLifespan(time.Minute*30)); err != nil {
		return nil, s.handleError(w, r, f, pid, nil, err)
	}

	f.Active = s.ID()
	if err = s.d.LoginFlowPersister().UpdateLoginFlow(ctx, f); err != nil {
		return nil, s.handleError(w, r, f, pid, nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Could not update flow").WithDebug(err.Error())))
	}

	var up map[string]string
	if err := json.NewDecoder(bytes.NewBuffer(p.UpstreamParameters)).Decode(&up); err != nil {
		return nil, err
	}

	codeURL, err := getAuthRedirectURL(ctx, provider, f, state, up)
	if err != nil {
		return nil, s.handleError(w, r, f, pid, nil, err)
	}

	if x.IsJSONRequest(r) {
		s.d.Writer().WriteError(w, r, flow.NewBrowserLocationChangeRequiredError(codeURL))
	} else {
		http.Redirect(w, r, codeURL, http.StatusSeeOther)
	}

	return nil, errors.WithStack(flow.ErrCompletedByStrategy)
}
