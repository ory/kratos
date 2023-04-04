// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessages(t *testing.T) {
	require.NoError(t, validateAllMessages("../../text"))
}
