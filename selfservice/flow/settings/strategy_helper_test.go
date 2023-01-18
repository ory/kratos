// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package settings

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
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
