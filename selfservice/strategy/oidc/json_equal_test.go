// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonEqual(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		a, b     json.RawMessage
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "both empty",
			a:        json.RawMessage{},
			b:        json.RawMessage{},
			expected: true,
		},
		{
			name:     "one nil one empty",
			a:        nil,
			b:        json.RawMessage{},
			expected: true,
		},
		{
			name:     "one nil one non-empty",
			a:        nil,
			b:        json.RawMessage(`{}`),
			expected: false,
		},
		{
			name:     "identical objects",
			a:        json.RawMessage(`{"name":"alice","age":30}`),
			b:        json.RawMessage(`{"name":"alice","age":30}`),
			expected: true,
		},
		{
			name:     "same content different key order",
			a:        json.RawMessage(`{"name":"alice","age":30}`),
			b:        json.RawMessage(`{"age":30,"name":"alice"}`),
			expected: true,
		},
		{
			name:     "same content different whitespace",
			a:        json.RawMessage(`{"name": "alice"}`),
			b:        json.RawMessage(`{"name":"alice"}`),
			expected: true,
		},
		{
			name:     "different values",
			a:        json.RawMessage(`{"name":"alice"}`),
			b:        json.RawMessage(`{"name":"bob"}`),
			expected: false,
		},
		{
			name:     "nested objects equal",
			a:        json.RawMessage(`{"user":{"name":"alice","groups":["admin","user"]}}`),
			b:        json.RawMessage(`{"user":{"groups":["admin","user"],"name":"alice"}}`),
			expected: true,
		},
		{
			name:     "nested objects different",
			a:        json.RawMessage(`{"user":{"groups":["admin"]}}`),
			b:        json.RawMessage(`{"user":{"groups":["admin","user"]}}`),
			expected: false,
		},
		{
			name:     "extra key",
			a:        json.RawMessage(`{"name":"alice"}`),
			b:        json.RawMessage(`{"name":"alice","age":30}`),
			expected: false,
		},
		{
			name:     "invalid json a",
			a:        json.RawMessage(`{invalid`),
			b:        json.RawMessage(`{"name":"alice"}`),
			expected: false,
		},
		{
			name:     "invalid json b",
			a:        json.RawMessage(`{"name":"alice"}`),
			b:        json.RawMessage(`{invalid`),
			expected: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, jsonEqual(tc.a, tc.b))
		})
	}
}
