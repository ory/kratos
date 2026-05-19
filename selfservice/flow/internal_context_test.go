// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/sqlxx"
)

type fakeInternalContexter struct {
	ic sqlxx.JSONRawMessage
}

func (f *fakeInternalContexter) EnsureInternalContext() {
	if len(f.ic) == 0 {
		f.ic = sqlxx.JSONRawMessage("{}")
	}
}
func (f *fakeInternalContexter) GetInternalContext() sqlxx.JSONRawMessage  { return f.ic }
func (f *fakeInternalContexter) SetInternalContext(b sqlxx.JSONRawMessage) { f.ic = b }

func TestRequestBaseURLInternalContext(t *testing.T) {
	t.Run("round-trips through InternalContext", func(t *testing.T) {
		f := &fakeInternalContexter{}
		require.NoError(t, SetRequestBaseURL(f, "http://localhost:4000"))
		assert.Equal(t, "http://localhost:4000", GetRequestBaseURL(f))
		assert.Contains(t, string(f.GetInternalContext()), InternalContextKeyRequestBaseURL)
	})

	t.Run("empty input is a no-op and leaves no key", func(t *testing.T) {
		f := &fakeInternalContexter{ic: sqlxx.JSONRawMessage("{}")}
		require.NoError(t, SetRequestBaseURL(f, ""))
		assert.Empty(t, GetRequestBaseURL(f))
		assert.Equal(t, "{}", string(f.GetInternalContext()))
	})

	t.Run("oversized input is rejected (row-bloat guard)", func(t *testing.T) {
		f := &fakeInternalContexter{}
		require.NoError(t, SetRequestBaseURL(f, "https://"+strings.Repeat("a", 8193)))
		assert.Empty(t, GetRequestBaseURL(f))
	})

	t.Run("missing key returns empty", func(t *testing.T) {
		assert.Empty(t, GetRequestBaseURL(&fakeInternalContexter{ic: sqlxx.JSONRawMessage(`{"other":"x"}`)}))
	})
}
