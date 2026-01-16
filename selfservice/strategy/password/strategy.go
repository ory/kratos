// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ory/kratos/x/nosurfx"

	"github.com/pkg/errors"

	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/httpx"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
)

var (
	_ login.Strategy                    = (*Strategy)(nil)
	_ registration.Strategy             = (*Strategy)(nil)
	_ identity.ActiveCredentialsCounter = (*Strategy)(nil)
)

type dependencies interface {
	logrusx.Provider
	httpx.WriterProvider
	nosurfx.CSRFTokenGeneratorProvider
	nosurfx.CSRFProvider
	httpx.ClientProvider
	otelx.Provider
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

type Strategy struct{ d dependencies }

func NewStrategy(d dependencies) *Strategy { return &Strategy{d: d} }

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
