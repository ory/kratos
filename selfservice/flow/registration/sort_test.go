// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"context"
	"embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/ui/node"
	"github.com/ory/x/snapshotx"
)

//go:embed fixtures/sort/*
var sortFixtures embed.FS

func TestSortNodes(t *testing.T) {
	ctx := context.Background()

	// TODO add more test cases.
	entries, err := sortFixtures.ReadDir("fixtures/sort")
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		t.Run("case="+entry.Name(), func(t *testing.T) {
			toSort, err := sortFixtures.ReadFile("fixtures/sort/" + entry.Name())
			require.NoError(t, err)

			var n node.Nodes
			require.NoError(t, json.Unmarshal(toSort, &n))
			require.NoError(t, SortNodes(ctx, n, "file://fixtures/sort.schema.json"))

			snapshotx.SnapshotT(t, n)
		})
	}
}
