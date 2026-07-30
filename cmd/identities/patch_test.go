// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities_test

import (
	"context"
	"testing"

	"github.com/ory/kratos/cmd/identities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

func TestPatchCmd(t *testing.T) {
	c := identities.NewPatchIdentityCmd()
	reg := setup(t, c)

	t.Run("case=patches successfully", func(t *testing.T) {
		// create identity to patch
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		require.NoError(t, reg.Persister().CreateIdentity(context.Background(), i))
		assert.Equal(t, true, i.IsActive())

		// patch identity state to inactive
		stdOut := execNoErr(t, c, i.ID.String(), "--op", "replace", "--value", "inactive", "--path", "/state")
		assert.Contains(t, stdOut, "state_changed_at")

		// expect identity to be patched
		pi, err := reg.Persister().GetIdentity(context.Background(), i.ID, identity.ExpandEverything)
		assert.NoError(t, err)
		assert.Equal(t, false, pi.IsActive())

	})

	t.Run("case=patches three identities", func(t *testing.T) {
		is, ids := makeIdentities(t, reg, 3)

		for _, i := range is {
			assert.Equal(t, true, i.IsActive())
		}

		stdOut := execNoErr(t, c, ids[0], ids[1], ids[2], "--op", "replace", "--value", "inactive", "--path", "/state")
		assert.Contains(t, stdOut, "state_changed_at")

		for _, i := range is {
			pi, err := reg.Persister().GetIdentity(context.Background(), i.ID, identity.ExpandEverything)
			assert.NoError(t, err)
			assert.Equal(t, false, pi.IsActive())
		}
	})

	t.Run("case=fails with unknown ID", func(t *testing.T) {
		stdErr := execErr(t, c, x.NewUUID().String())

		assert.Contains(t, stdErr, "Unable to locate the resource", stdErr)
	})

}
