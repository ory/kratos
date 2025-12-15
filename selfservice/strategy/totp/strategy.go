// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package totp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ory/kratos/x/nosurfx"

	"github.com/pkg/errors"
	"github.com/pquerna/otp"

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
	_ login.AAL2FormHydrator            = new(Strategy)
	_ settings.Strategy                 = new(Strategy)
	_ identity.ActiveCredentialsCounter = new(Strategy)
)

type totpStrategyDependencies interface {
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

	session.HandlerProvider
	session.ManagementProvider
	session.PersistenceProvider
}

type Strategy struct {
	d  totpStrategyDependencies
	hd *decoderx.HTTP
}

func NewStrategy(d totpStrategyDependencies) *Strategy {
	return &Strategy{
		d:  d,
		hd: decoderx.NewHTTP(),
	}
}

func (s *Strategy) CountActiveFirstFactorCredentials(_ context.Context, _ map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	return 0, nil
}

func (s *Strategy) CountActiveMultiFactorCredentials(_ context.Context, cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	for _, c := range cc {
		if c.Type == s.ID() && len(c.Config) > 0 {
			var conf identity.CredentialsTOTPConfig
			if err = json.Unmarshal(c.Config, &conf); err != nil {
				return 0, errors.WithStack(err)
			}

			_, err := otp.NewKeyFromURL(conf.TOTPURL)
			if len(conf.TOTPURL) > 0 && err == nil {
				count++
			}
		}
	}
	return
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsTypeTOTP
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.TOTPGroup
}

func (s *Strategy) CompletedAuthenticationMethod(ctx context.Context) session.AuthenticationMethod {
	return session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    identity.AuthenticatorAssuranceLevel2,
	}
}

func (s *Strategy) PopulateLoginMethodSecondFactor(r *http.Request, f *login.Flow) error {
	return s.PopulateLoginMethod(r, identity.AuthenticatorAssuranceLevel2, f)
}

func (s *Strategy) PopulateLoginMethodSecondFactorRefresh(r *http.Request, f *login.Flow) error {
	return s.PopulateLoginMethod(r, identity.AuthenticatorAssuranceLevel2, f)
}
