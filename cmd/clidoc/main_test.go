// Copyright © 2022 Ory Corp

package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessages(t *testing.T) {
	require.NoError(t, validateAllMessages("../../text"))
}
