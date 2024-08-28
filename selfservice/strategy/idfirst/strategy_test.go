// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package idfirst_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/x"
	"github.com/ory/x/sqlxx"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/idfirst"
	"github.com/ory/kratos/ui/node"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/identity"
)

func TestCountActiveFirstFactorCredentials(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	s := idfirst.NewStrategy(reg)
	cc := make(map[identity.CredentialsType]identity.Credentials)

	count, err := s.CountActiveFirstFactorCredentials(cc)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestCountActiveMultiFactorCredentials(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	s := idfirst.NewStrategy(reg)
	cc := make(map[identity.CredentialsType]identity.Credentials)

	count, err := s.CountActiveMultiFactorCredentials(cc)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestCompletedAuthenticationMethod(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	s := idfirst.NewStrategy(reg)
	ctx := context.Background()

	method := s.CompletedAuthenticationMethod(ctx)
	assert.Equal(t, s.ID(), method.Method)
	assert.Equal(t, identity.NoAuthenticatorAssuranceLevel, method.AAL)
}

func TestNodeGroup(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	s := idfirst.NewStrategy(reg)

	group := s.NodeGroup()
	assert.Equal(t, node.IdentifierFirstGroup, group)
}

func createIdentity(ctx context.Context, reg *driver.RegistryDefault, t *testing.T, identifier, password string) *identity.Identity {
	p, _ := reg.Hasher(ctx).Generate(context.Background(), []byte(password))
	iId := x.NewUUID()
	id := &identity.Identity{
		ID:     iId,
		Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, identifier)),
		Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{identifier},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
			},
		},
		VerifiableAddresses: []identity.VerifiableAddress{
			{
				ID:         x.NewUUID(),
				Value:      identifier,
				Verified:   false,
				CreatedAt:  time.Now(),
				IdentityID: iId,
			},
		},
	}
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), id))
	return id
}
