// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestParseRawMessageOrEmpty(t *testing.T) {
	for _, tc := range []struct {
		input  json.RawMessage
		expect map[string]interface{}
		err    any
	}{
		{
			input: json.RawMessage("invalid json"),
			err:   "invalid character 'i' looking for beginning of value",
		},
		{
			input:  json.RawMessage(""),
			expect: map[string]interface{}{},
		},
		{
			input: json.RawMessage(`{"foo": "bar"}`),
			expect: map[string]interface{}{
				"foo": "bar",
			},
		},
		{
			input: json.RawMessage(`{"foo": "b`),
			err:   "unexpected end of JSON input",
		},
	} {
		t.Run(fmt.Sprintf("with input '%s'", tc.input), func(t *testing.T) {
			m, err := x.ParseRawMessageOrEmpty(tc.input)
			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err, err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expect, m)
		})
	}
}
