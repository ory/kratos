// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package idfirst_test

import (
	"context"
	"fmt"
	"testing"

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

func createIdentity(ctx context.Context, reg *driver.RegistryDefault, t *testing.T, email string, withPassword bool, withCode bool) *identity.Identity {
	iId := x.NewUUID()
	id := &identity.Identity{
		ID:          iId,
		Traits:      identity.Traits(fmt.Sprintf(`{ "email": "%s" }`, email)),
		Credentials: map[identity.CredentialsType]identity.Credentials{},
		VerifiableAddresses: []identity.VerifiableAddress{
			{
				Value:    email,
				Verified: true,
				Status:   identity.VerifiableAddressStatusCompleted,
			},
		},
	}
	if withPassword {
		p, _ := reg.Hasher(ctx).Generate(context.Background(), []byte("password"))
		id.Credentials[identity.CredentialsTypePassword] = identity.Credentials{
			Type:        identity.CredentialsTypePassword,
			Identifiers: []string{email},
			Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
		}
	}
	if withCode {
		id.Credentials[identity.CredentialsTypeCodeAuth] = identity.Credentials{
			Type:        identity.CredentialsTypeCodeAuth,
			Identifiers: []string{email}, Config: sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"` + email + `"}]}`),
		}
	}
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), id))
	return id
}

func createIdentityWithAllMethods(ctx context.Context, reg *driver.RegistryDefault, t *testing.T, email string) *identity.Identity {
	id := x.NewUUID()
	pwd, _ := reg.Hasher(ctx).Generate(context.Background(), []byte("password"))
	newIdentity := &identity.Identity{
		ID:     id,
		Traits: identity.Traits(fmt.Sprintf(`{ "email": "%s" }`, email)),
		Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{email},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(pwd) + `"}`),
			},
			identity.CredentialsTypeCodeAuth: {
				Type:        identity.CredentialsTypeCodeAuth,
				Identifiers: []string{email},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"` + email + `"}]}`),
			},
			identity.CredentialsTypeOIDC: {
				Type:        identity.CredentialsTypeOIDC,
				Identifiers: []string{email},
				Config:      sqlxx.JSONRawMessage(`{"some" : "secret"}`),
			},
			identity.CredentialsTypeSAML: {
				Type:        identity.CredentialsTypeSAML,
				Identifiers: []string{email},
				Config:      sqlxx.JSONRawMessage(`{"saml" : "secret"}`),
			},
			identity.CredentialsTypeWebAuthn: {
				Type:        identity.CredentialsTypeWebAuthn,
				Identifiers: []string{email},
				Config:      sqlxx.JSONRawMessage(`{"some" : "secret", "user_handle": "rVIFaWRcTTuQLkXFmQWpgA=="}`),
			},
			identity.CredentialsTypePasskey: {
				Type:        identity.CredentialsTypePasskey,
				Identifiers: []string{email},
				Config:      []byte(`{"credentials":[{"id":"Zm9vZm9v","display_name":"foo"},{"id":"YmFyYmFy","display_name":"bar"}]}`),
			},
		},
		VerifiableAddresses: []identity.VerifiableAddress{
			{
				Value:    email,
				Verified: true,
				Status:   identity.VerifiableAddressStatusCompleted,
			},
		},
	}
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), newIdentity))
	return newIdentity
}
