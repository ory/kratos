// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package idfirst

import (
	"context"

	"github.com/go-playground/validator/v10"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/ui/node"
	"github.com/ory/kratos/x"
	"github.com/ory/x/decoderx"
)

type dependencies interface {
	x.LoggingProvider
	x.WriterProvider
	x.CSRFTokenGeneratorProvider
	x.CSRFProvider
	x.TracingProvider

	config.Provider

	identity.PrivilegedPoolProvider
	login.StrategyProvider
	login.FlowPersistenceProvider
}

type Strategy struct {
	d  dependencies
	v  *validator.Validate
	hd *decoderx.HTTP
}

func NewStrategy(d any) *Strategy {
	return &Strategy{
		d:  d.(dependencies),
		v:  validator.New(),
		hd: decoderx.NewHTTP(),
	}
}

func (s *Strategy) CountActiveFirstFactorCredentials(cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	return 0, nil
}

func (s *Strategy) CountActiveMultiFactorCredentials(cc map[identity.CredentialsType]identity.Credentials) (count int, err error) {
	return 0, nil
}

func (s *Strategy) ID() identity.CredentialsType {
	return identity.CredentialsType(node.IdentifierFirstGroup)
}

func (s *Strategy) CompletedAuthenticationMethod(ctx context.Context) session.AuthenticationMethod {
	return session.AuthenticationMethod{
		Method: s.ID(),
		AAL:    identity.NoAuthenticatorAssuranceLevel,
	}
}

func (s *Strategy) NodeGroup() node.UiNodeGroup {
	return node.IdentifierFirstGroup
}
