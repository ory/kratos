// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ory/kratos/x/nosurfx"

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

type strategyDependencies interface {
	x.LoggingProvider
	x.WriterProvider
	nosurfx.CSRFTokenGeneratorProvider
	nosurfx.CSRFProvider
	x.TracingProvider

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
	identity.ActiveCredentialsCounterStrategyProvider
	identity.ManagementProvider

	session.HandlerProvider
	session.ManagementProvider
}

var (
	_ login.Strategy                    = new(Strategy)
	_ registration.Strategy             = new(Strategy)
	_ identity.ActiveCredentialsCounter = new(Strategy)
)

type Strategy struct {
	d  strategyDependencies
	hd *decoderx.HTTP
}

func NewStrategy(d strategyDependencies) *Strategy {
	return &Strategy{
		d:  d,
		hd: decoderx.NewHTTP(),
	}
}

func (*Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypePasskey
}

func (*Strategy) NodeGroup() node.UiNodeGroup {
	return node.PasskeyGroup
}

func (s *Strategy) CompletedAuthenticationMethod(context.Context) session.AuthenticationMethod {
	return session.AuthenticationMethod{
		Method: identity.CredentialsTypePasskey,
		AAL:    identity.AuthenticatorAssuranceLevel1,
	}
}

func (s *Strategy) CountActiveMultiFactorCredentials(_ context.Context, _ map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	return 0, nil
}

func (s *Strategy) CountActiveFirstFactorCredentials(_ context.Context, cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	return s.countCredentials(cc)
}

func (s *Strategy) countCredentials(cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	for _, c := range cc {
		if c.Type == s.ID() && len(c.Config) > 0 && len(strings.Join(c.Identifiers, "")) > 0 {
			var conf identity.CredentialsWebAuthnConfig
			if err = json.Unmarshal(c.Config, &conf); err != nil {
				return 0, errors.WithStack(err)
			}
			count += len(conf.Credentials)
		}
	}
	return
}
