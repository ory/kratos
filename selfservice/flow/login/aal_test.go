package login_test

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckAAL(t *testing.T) {
	f := &login.Flow{RequestedAAL: identity.AuthenticatorAssuranceLevel1}
	assert.NoError(t, login.CheckAAL(f, identity.AuthenticatorAssuranceLevel1))
	assert.ErrorIs(t, login.CheckAAL(f, identity.AuthenticatorAssuranceLevel2), flow.ErrStrategyNotResponsible)
}
