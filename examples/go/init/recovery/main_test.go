package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecovery(t *testing.T) {
	flow := initRecovery()
	require.NotEmpty(t, flow.Id)
}
