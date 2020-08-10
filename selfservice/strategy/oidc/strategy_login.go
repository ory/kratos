package oidc

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

var _ login.Strategy = new(Strategy)

func (s *Strategy) RegisterLoginRoutes(r *x.RouterPublic) {
	s.setRoutes(r)
}

func (s *Strategy) PopulateLoginMethod(r *http.Request, sr *login.Flow) error {
	config, err := s.populateMethod(r, sr.ID)
	if err != nil {
		return err
	}
	sr.Methods[s.ID()] = &login.FlowMethod{Method: s.ID(),
		Config: &login.FlowMethodConfig{FlowMethodConfigurator: config}}
	return nil
}

func (s *Strategy) processLogin(w http.ResponseWriter, r *http.Request, a *login.Flow, claims *Claims, provider Provider, container *authCodeContainer) {
	i, c, err := s.d.PrivilegedIdentityPool().FindByCredentialsIdentifier(r.Context(), identity.CredentialsTypeOIDC, uid(provider.Config().ID, claims.Subject))
	if err != nil {
		if errors.Is(err, herodot.ErrNotFound) {
			// If no account was found we're "manually" creating a new registration request and redirecting the browser
			// to that endpoint.

			// That will execute the "pre registration" hook which allows to e.g. disallow this request. The registration
			// ui however will NOT be shown, instead the user is directly redirected to the auth path. That should then
			// do a silent re-request. While this might be a bit excessive from a network perspective it should usually
			// happen without any downsides to user experience as the request has already been authorized and should
			// not need additional consent/login.

			// This is kinda hacky but the only way to ensure seamless login/registration flows when using OIDC.

			s.d.Logger().WithField("provider", provider.Config().ID).WithField("subject", claims.Subject).Debug("Received successful OpenID Connect callback but user is not registered. Re-initializing registration flow now.")
			aa, err := s.d.RegistrationHandler().NewRegistrationRequest(w, r)
			if err != nil {
				s.handleError(w, r, a.GetID(), provider.Config().ID, nil, err)
				return
			}

			s.processRegistration(w, r, aa, claims, provider, container)
			return
		}

		s.handleError(w, r, a.GetID(), provider.Config().ID, nil, err)
		return
	}

	var o CredentialsConfig
	if err := json.NewDecoder(bytes.NewBuffer(c.Config)).Decode(&o); err != nil {
		s.handleError(w, r, a.GetID(), provider.Config().ID, nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("The password credentials could not be decoded properly").WithDebug(err.Error())))
		return
	}

	for _, c := range o.Providers {
		if c.Subject == claims.Subject && c.Provider == provider.Config().ID {
			if err = s.d.LoginHookExecutor().PostLoginHook(w, r, identity.CredentialsTypeOIDC, a, i); err != nil {
				s.handleError(w, r, a.GetID(), provider.Config().ID, nil, err)
				return
			}
			return
		}
	}

	s.handleError(w, r, a.GetID(), provider.Config().ID, nil, errors.WithStack(herodot.ErrInternalServerError.WithReason("Unable to find matching OpenID Connect Credentials.").WithDebugf(`Unable to find credentials that match the given provider "%s" and subject "%s".`, provider.Config().ID, claims.Subject)))
}
