package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistration(t *testing.T) {
	flow := initRegistration()
	require.NotEmpty(t, flow.Id)
}
