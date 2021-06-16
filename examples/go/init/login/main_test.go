package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	flow := initLogin()
	require.NotEmpty(t, flow.Id)
}
