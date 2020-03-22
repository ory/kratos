package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/selfservice/hook"
)

func (m *RegistryDefault) getHooks(credentialsType string, configs []configuration.SelfServiceHook) []interface{} {
	var i []interface{}

	for _, h := range configs {
		switch h.Job {
		case hook.KeyVerify:
			i = append(
				i,
				hook.NewVerifier(m),
			)
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
			var rc struct {
				R string `json:"default_redirect_url"`
				A bool   `json:"allow_user_defined_redirect"`
			}

			if err := json.NewDecoder(bytes.NewBuffer(h.Config)).Decode(&rc); err != nil {
				m.l.WithError(err).
					WithField("type", credentialsType).
					WithField("hook", h.Job).
					WithField("config", fmt.Sprintf("%s", h.Config)).
					Errorf("The after hook is misconfigured.")
				continue
			}

			rcr, err := url.ParseRequestURI(rc.R)
			if err != nil {
				m.l.WithError(err).
					WithField("type", credentialsType).
					WithField("hook", h.Job).
					WithField("config", fmt.Sprintf("%s", h.Config)).
					Errorf("The after hook is misconfigured.")
				continue
			}

			i = append(
				i,
				hook.NewRedirector(
					func() *url.URL {
						return rcr
					},
					m.c.WhitelistedReturnToDomains,
					func() bool {
						return rc.A
					},
				),
			)
		default:
			m.l.
				WithField("type", credentialsType).
				WithField("hook", h.Job).
				Errorf("A unknown hook was requested and can therefore not be used")
		}
	}

	return i
}
