// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identities_test

import (
	"context"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/ory/kratos/cmd/identities"

	"github.com/ory/x/cmdx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
)

func TestListCmd(t *testing.T) {
	reg, cmd := setup(t, identities.NewListIdentitiesCmd)

	var deleteIdentities = func(t *testing.T, is []*identity.Identity) {
		for _, i := range is {
			require.NoError(t, reg.Persister().DeleteIdentity(context.Background(), i.ID))
		}
	}

	t.Run("case=lists all identities with default pagination", func(t *testing.T) {
		is, ids := makeIdentities(t, reg, 5)
		t.Cleanup(func() {
			deleteIdentities(t, is)
		})

		stdOut := cmd.ExecNoErr(t, "--"+cmdx.FlagQuiet)

		for _, i := range ids {
			assert.Contains(t, stdOut, i)
		}
	})

	t.Run("case=lists all identities with pagination", func(t *testing.T) {
		is, ids := makeIdentities(t, reg, 10)
		t.Cleanup(func() {
			deleteIdentities(t, is)
		})

		first := cmd.ExecNoErr(t, "--format", "json-pretty", "--page-size", "2")
		nextPageToken := gjson.Get(first, "next_page_token").String()
		actualIDs := gjson.Get(first, "identities.#.id").Array()
		for nextPageToken != "" {
			next := cmd.ExecNoErr(t, "--page-size", "2", "--page-token", nextPageToken)
			actualIDs = append(actualIDs, gjson.Get(next, "identities.#.id").Array()...)
			nextPageToken = gjson.Get(next, "next_page_token").String()
		}

		actualIDsString := make([]string, len(actualIDs))
		for i, id := range actualIDs {
			actualIDsString[i] = id.Str
		}

		assert.ElementsMatch(t, ids, actualIDsString)
	})
}
