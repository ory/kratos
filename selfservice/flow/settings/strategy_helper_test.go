package settings

import (
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetIdentityToUpdate(t *testing.T) {
	c := new(UpdateContext)
	_, err := c.GetIdentityToUpdate()
	require.Error(t, err)

	expected := &identity.Identity{ID: x.NewUUID()}
	c.UpdateIdentity(expected)

	actual, err := c.GetIdentityToUpdate()
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
