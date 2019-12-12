package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"net/url"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/hook"
)

func (m *RegistryDefault) hooksPostRegistration(credentialsType identity.CredentialsType, configs []configuration.SelfServiceHook) []registration.PostHookExecutor {
	var i []registration.PostHookExecutor

	for _, h := range configs {
		switch h.Run {
		case hook.KeySessionIssuer:
			i = append(
				i,
				hook.NewSessionIssuer(m),
			)
		case hook.KeyRedirector:
			redirector, err := newRedirector(m, h, credentialsType)
			if err != nil {
				continue
			}
			i = append(
				i,
				redirector,
			)
		default:
			m.l.
				WithField("type", credentialsType).
				WithField("hook", h.Run).
				Errorf("A unknown post registration hook was requested and can therefore not be used.")
		}
	}

	return i
}

func (m *RegistryDefault) hooksPostLogin(credentialsType identity.CredentialsType, configs []configuration.SelfServiceHook) []login.PostHookExecutor {
	var i []login.PostHookExecutor

	for _, h := range configs {
		switch h.Run {
		case hook.KeySessionIssuer:
			i = append(
				i,
				hook.NewSessionIssuer(m),
			)
		case hook.KeySessionDestroyer:
			i = append(
				i,
				hook.NewSessionDestroyer(m),
			)
		case hook.KeyRedirector:
			redirector, err := newRedirector(m, h, credentialsType)
			if err != nil {
				continue
			}
			i = append(
				i,
				redirector,
			)
		default:
			m.l.
				WithField("type", credentialsType).
				WithField("hook", h.Run).
				Errorf("A unknown post login hook was requested and can therefore not be used.")
		}
	}

	return i
}

func newRedirector(m *RegistryDefault, h configuration.SelfServiceHook, credentialsType identity.CredentialsType) (*hook.Redirector, error) {
	var rc struct {
		R string `json:"default_redirect_url"`
		A bool   `json:"allow_user_defined_redirect"`
	}

	if err := json.NewDecoder(bytes.NewBuffer(h.Config)).Decode(&rc); err != nil {
		m.l.WithError(err).
			WithField("type", credentialsType).
			WithField("hook", h.Run).
			WithField("config", fmt.Sprintf("%s", h.Config)).
			Errorf("The after hook is misconfigured.")
		return nil, err
	}

	rcr, err := url.ParseRequestURI(rc.R)
	if err != nil {
		m.l.WithError(err).
			WithField("type", credentialsType).
			WithField("hook", h.Run).
			WithField("config", fmt.Sprintf("%s", h.Config)).
			Errorf("The after hook is misconfigured.")
		return nil, err
	}

	return hook.NewRedirector(
		func() *url.URL {
			return rcr
		},
		m.c.WhitelistedReturnToDomains,
		func() bool {
			return rc.A
		},
	), nil
}
