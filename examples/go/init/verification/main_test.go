package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerification(t *testing.T) {
	flow := initVerification()
	require.NotEmpty(t, flow.Id)
}
