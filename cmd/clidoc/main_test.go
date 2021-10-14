package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMessages(t *testing.T) {
	require.NoError(t, validateAllMessages("../../text"))
}
