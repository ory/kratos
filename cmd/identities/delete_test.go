// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/ory/kratos/cmd/identities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
	"github.com/ory/x/sqlcon"
)

func TestDeleteCmd(t *testing.T) {
	c := identities.NewDeleteIdentityCmd()
	reg := setup(t, c)

	t.Run("case=deletes successfully", func(t *testing.T) {
		// create identity to delete
		i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
		require.NoError(t, reg.Persister().CreateIdentity(context.Background(), i))

		stdOut := execNoErr(t, c, i.ID.String())

		// expect ID and no error
		assert.Equal(t, i.ID.String(), gjson.Parse(stdOut).String())

		// expect identity to be deleted
		_, err := reg.Persister().GetIdentity(context.Background(), i.ID)
		assert.True(t, errors.Is(err, sqlcon.ErrNoRows))
	})

	t.Run("case=deletes three identities", func(t *testing.T) {
		is, ids := makeIdentities(t, reg, 3)

		stdOut := execNoErr(t, c, ids...)

		assert.Equal(t, `["`+strings.Join(ids, "\",\"")+"\"]\n", stdOut)

		for _, i := range is {
			_, err := reg.Persister().GetIdentity(context.Background(), i.ID)
			assert.Error(t, err)
		}
	})

	t.Run("case=fails with unknown ID", func(t *testing.T) {
		stdErr := execErr(t, c, x.NewUUID().String())

		assert.Contains(t, stdErr, "Unable to locate the resource", stdErr)
	})
}
