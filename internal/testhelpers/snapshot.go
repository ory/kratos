// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"encoding/json"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"
)

func SnapshotTExcept(t *testing.T, actual interface{}, except []string) {
	t.Helper()
	compare, err := json.MarshalIndent(actual, "", "  ")
	require.NoError(t, err, "%+v", actual)
	for _, e := range except {
		compare, err = sjson.DeleteBytes(compare, e)
		require.NoError(t, err, "%s", e)
	}

	cupaloy.New(
		cupaloy.CreateNewAutomatically(true),
		cupaloy.FailOnUpdate(true),
		cupaloy.SnapshotFileExtension(".json"),
	).SnapshotT(t, compare)
}
