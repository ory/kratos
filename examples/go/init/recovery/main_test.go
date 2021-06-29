package main

import (
	"testing"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/kratos/examples/go/pkg"

	"github.com/stretchr/testify/require"
)

func TestRecovery(t *testing.T) {
	publicURL, _ := testhelpers.StartE2EServer(t, "../../pkg/stub/kratos.yaml")
	client = pkg.NewSDKForSelfHosted(publicURL)

	flow := initRecovery()
	require.NotEmpty(t, flow.Id)
}
