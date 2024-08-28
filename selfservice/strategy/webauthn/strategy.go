// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package webauthn

import (
	"context"
	"encoding/json"
	"strings"

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

var (
	_ login.Strategy                    = new(Strategy)
	_ settings.Strategy                 = new(Strategy)
	_ identity.ActiveCredentialsCounter = new(Strategy)
)

type webauthnStrategyDependencies interface {
	x.LoggingProvider
	x.WriterProvider
	x.CSRFTokenGeneratorProvider
	x.CSRFProvider
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

type Strategy struct {
	d  webauthnStrategyDependencies
	hd *decoderx.HTTP
}

func NewStrategy(d any) *Strategy {
	return &Strategy{
		d:  d.(webauthnStrategyDependencies),
		hd: decoderx.NewHTTP(),
	}
}

func (s *Strategy) CountActiveMultiFactorCredentials(_ context.Context, cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	return s.countCredentials(cc, false)
}

func (s *Strategy) CountActiveFirstFactorCredentials(_ context.Context, cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	return s.countCredentials(cc, true)
}

func (s *Strategy) countCredentials(cc map[identity.CredentialsType]identity.Credentials, onlyPasswordlessCredentials bool) (count int, err error) {
	for _, c := range cc {
		if c.Type == s.ID() && len(c.Config) > 0 {
			var conf identity.CredentialsWebAuthnConfig
			if err = json.Unmarshal(c.Config, &conf); err != nil {
				return 0, errors.WithStack(err)
			}

			for _, cred := range conf.Credentials {
				if cred.IsPasswordless && len(strings.Join(c.Identifiers, "")) == 0 {
					// If this is a passwordless credential, it will only work if the identifier is set, as
					// we use the identifier to look up the identity. If the identifier is not set, we can
					// assume that the user can't sign in using this method.
					continue
				}

				if cred.IsPasswordless != onlyPasswordlessCredentials {
					continue
				}

				// If the credential is passwordless and we require passwordless credentials, or if the credential is not
				// passwordless and we require non-passwordless credentials, we count it.
				count++
			}
		}
	}
	return
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeWebAuthn
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.WebAuthnGroup
}

func (s *Strategy) CompletedAuthenticationMethod(ctx context.Context) session.AuthenticationMethod {
	aal := identity.AuthenticatorAssuranceLevel1
	if !s.d.Config().WebAuthnForPasswordless(ctx) {
		aal = identity.AuthenticatorAssuranceLevel2
	}
	return session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    aal,
	}
}
