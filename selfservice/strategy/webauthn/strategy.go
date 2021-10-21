package webauthn

import (
	"context"
	"encoding/json"

	"github.com/duo-labs/webauthn/webauthn"

	"github.com/pkg/errors"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
)

var _ login.Strategy = new(Strategy)
var _ settings.Strategy = new(Strategy)
var _ identity.ActiveCredentialsCounter = new(Strategy)

type registrationStrategyDependencies interface {
	x.LoggingProvider
	x.WriterProvider
	x.CSRFTokenGeneratorProvider
	x.CSRFProvider

	config.Provider

	continuity.ManagementProvider

	errorx.ManagementProvider
	hash.HashProvider

	registration.HandlerProvider
	registration.HooksProvider
	registration.ErrorHandlerProvider
	registration.HookExecutorProvider
	registration.FlowPersistenceProvider

	login.HooksProvider
	login.ErrorHandlerProvider
	login.HookExecutorProvider
	login.FlowPersistenceProvider
	login.HandlerProvider

	settings.FlowPersistenceProvider
	settings.HookExecutorProvider
	settings.HooksProvider
	settings.ErrorHandlerProvider

	identity.PrivilegedPoolProvider
	identity.ValidationProvider

	session.HandlerProvider
	session.ManagementProvider
}

type Strategy struct {
	d  registrationStrategyDependencies
	hd *decoderx.HTTP
}

func NewStrategy(d registrationStrategyDependencies) *Strategy {
	return &Strategy{
		d:  d,
		hd: decoderx.NewHTTP(),
	}
}

func (s *Strategy) CountActiveCredentials(cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	for _, c := range cc {
		if c.Type == s.ID() && len(c.Config) > 0 {
			var conf CredentialsConfig
			if err = json.Unmarshal(c.Config, &conf); err != nil {
				return 0, errors.WithStack(err)
			}

			if len(conf.Credentials) > 0 {
				count++
			}
		}
	}
	return
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeWebAuthn
}

func (s *Strategy) NodeGroup() node.Group {
	return node.WebAuthnGroup
}

func (s *Strategy) newWebAuthn(ctx context.Context) (*webauthn.WebAuthn, error) {
	c := s.d.Config(ctx)
	web, err := webauthn.New(c.WebAuthnConfig())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return web, nil
}
