// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSalesforceUpdatedAtWorkaround(t *testing.T) {
	actual, err := authZeroUpdatedAtWorkaround([]byte("{}"))
	require.NoError(t, err)
	assert.Equal(t, "{}", string(actual))

	actual, err = authZeroUpdatedAtWorkaround([]byte(`{"updated_at":1234}`))
	require.NoError(t, err)
	assert.Equal(t, `{"updated_at":1234}`, string(actual))

	timestamp := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	input, err := json.Marshal(map[string]interface{}{
		"updated_at": timestamp,
	})
	require.NoError(t, err)
	actual, err = authZeroUpdatedAtWorkaround(input)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(`{"updated_at":%d}`, timestamp.Unix()), string(actual))
}
