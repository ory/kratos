// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package jsonnetsecure_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/jsonnetsecure"
)

func TestProcessPoolVMAcceptsInputLargerThanDefaultScannerBuffer(t *testing.T) {
	const defaultScannerBuffer = 64 * 1024
	largeValue := strings.Repeat("a", 2*defaultScannerBuffer)

	vm, err := jsonnetsecure.NewTestProvider(t).JsonnetVM(context.Background())
	require.NoError(t, err)

	vm.ExtCode("claims", `{"name": "`+largeValue+`"}`)
	out, err := vm.EvaluateAnonymousSnippet("test.jsonnet", `{ length: std.length(std.extVar("claims").name) }`)
	require.NoError(t, err)

	var result struct {
		Length int `json:"length"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &result))
	assert.Equal(t, len(largeValue), result.Length)
}
