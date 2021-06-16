package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSettings(t *testing.T) {
	flow := initSettings()
	require.NotEmpty(t, flow.Id)
}
