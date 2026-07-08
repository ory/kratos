// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package deviceauthn_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice/strategy/deviceauthn"
)

func TestKeyVersion2RoundTrip(t *testing.T) {
	t.Parallel()
	k := deviceauthn.Key{
		Version:          2,
		ClientKeyID:      "abc123",
		State:            deviceauthn.KeyStateConfirmed,
		UserVerification: deviceauthn.UserVerificationPIN,
		PIN: &deviceauthn.PINConfig{
			PINSecret:      "ciphertext-hex",
			FailedAttempts: 2,
		},
	}
	b, err := json.Marshal(k)
	require.NoError(t, err)

	var got deviceauthn.Key
	require.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, k, got)
}

func TestKeyVersion1ParsesAsLegacy(t *testing.T) {
	t.Parallel()
	legacy := []byte(`{"version":1,"client_key_id":"old","state":"confirmed"}`)
	var got deviceauthn.Key
	require.NoError(t, json.Unmarshal(legacy, &got))
	assert.Equal(t, 1, got.Version)
	assert.Nil(t, got.PIN)
	assert.Equal(t, deviceauthn.UserVerification(""), got.UserVerification)
}
