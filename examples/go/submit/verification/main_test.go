package main

import (
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

func TestVerification(t *testing.T) {
	flow := performVerification("dev+" + uuid.New().String() + "@ory.sh")
	require.NotEmpty(t, flow.Id)
}
