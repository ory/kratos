package idfirst_test

import (
	"context"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/idfirst"
	"github.com/ory/kratos/ui/node"
	"testing"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/session"
	"github.com/stretchr/testify/assert"
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

	method := s.CompletedAuthenticationMethod(ctx, session.AuthenticationMethods{})
	assert.Equal(t, s.ID(), method.Method)
	assert.Equal(t, identity.AuthenticatorAssuranceLevel1, method.AAL)
}

func TestNodeGroup(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	s := idfirst.NewStrategy(reg)

	group := s.NodeGroup()
	assert.Equal(t, node.IdentifierFirstGroup, group)
}
