// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func RequireJSONMarshal(t *testing.T, in interface{}) []byte {
	var b bytes.Buffer
	require.NoError(t, json.NewEncoder(&b).Encode(in))
	return b.Bytes()
}
