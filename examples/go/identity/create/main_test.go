package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/examples/go/pkg"
)

func TestFunc(t *testing.T) {
	client = pkg.TestClient(t)
	require.NotEmpty(t, createIdentity().Id)
}
