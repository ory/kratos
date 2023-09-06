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
	c := identities.NewListIdentitiesCmd()
	reg := setup(t, c)

	var deleteIdentities = func(t *testing.T, is []*identity.Identity) {
		for _, i := range is {
			require.NoError(t, reg.Persister().DeleteIdentity(context.Background(), i.ID))
		}
	}

	t.Run("case=lists all identities with default pagination", func(t *testing.T) {
		is, ids := makeIdentities(t, reg, 5)
		require.NoError(t, c.Flags().Set(cmdx.FlagQuiet, "true"))
		t.Cleanup(func() {
			require.NoError(t, c.Flags().Set(cmdx.FlagQuiet, "false"))
			deleteIdentities(t, is)
		})

		stdOut := cmdx.ExecNoErr(t, c)

		for _, i := range ids {
			assert.Contains(t, stdOut, i)
		}
	})

	t.Run("case=lists all identities with pagination", func(t *testing.T) {
		is, ids := makeIdentities(t, reg, 10)
		t.Cleanup(func() {
			deleteIdentities(t, is)
		})

		first := cmdx.ExecNoErr(t, c, "--format", "json-pretty", "--page-size", "2")
		nextPageToken := gjson.Get(first, "next_page_token").String()
		results := gjson.Get(first, "identities").Array()
		for nextPageToken != "" {
			next := cmdx.ExecNoErr(t, c, "--format", "json-pretty", "--page-size", "2", "--page-token", nextPageToken)
			results = append(results, gjson.Get(next, "identities").Array()...)
			nextPageToken = gjson.Get(next, "next_page_token").String()
		}

		assert.Len(t, results, len(ids))
		for _, expected := range ids {
			var found bool
			for _, actual := range results {
				if actual.Get("id").String() == expected {
					found = true
					break
				}
			}
			require.True(t, found, "could not find id: %s", expected)
		}
	})
}
