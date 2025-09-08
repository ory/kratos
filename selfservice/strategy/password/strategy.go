// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ory/kratos/x/nosurfx"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/jsonnetsecure"

	"github.com/ory/x/decoderx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var (
	_ login.Strategy                    = new(Strategy)
	_ registration.Strategy             = new(Strategy)
	_ identity.ActiveCredentialsCounter = new(Strategy)
)

type registrationStrategyDependencies interface {
	x.LoggingProvider
	x.WriterProvider
	nosurfx.CSRFTokenGeneratorProvider
	nosurfx.CSRFProvider
	x.HTTPClientProvider
	x.TracingProvider
	jsonnetsecure.VMProvider
	config.Provider
	continuity.ManagementProvider

	errorx.ManagementProvider
	ValidationProvider
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
	identity.ManagementProvider

	session.HandlerProvider
	session.ManagementProvider
}

type Strategy struct {
	d  registrationStrategyDependencies
	v  *validator.Validate
	hd *decoderx.HTTP
}

func NewStrategy(d registrationStrategyDependencies) *Strategy {
	return &Strategy{
		d:  d,
		v:  validator.New(),
		hd: decoderx.NewHTTP(),
	}
}

func (s *Strategy) CountActiveFirstFactorCredentials(ctx context.Context, cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	for _, c := range cc {
		if c.Type == s.ID() && len(c.Config) > 0 {
			var conf identity.CredentialsPassword
			if err = json.Unmarshal(c.Config, &conf); err != nil {
				return 0, errors.WithStack(err)
			}

			if len(strings.Join(c.Identifiers, "")) > 0 &&
				((s.d.Config().PasswordMigrationHook(ctx).Enabled && conf.UsePasswordMigrationHook) ||
					len(conf.HashedPassword) > 0) {
				count++
			}
		}
	}
	return
}

func (s *Strategy) CountActiveMultiFactorCredentials(_ context.Context, _ map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	return 0, nil
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypePassword
}

func (s *Strategy) CompletedAuthenticationMethod(_ context.Context) session.AuthenticationMethod {
	return session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    identity.AuthenticatorAssuranceLevel1,
	}
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.PasswordGroup
}
