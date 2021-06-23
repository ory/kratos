package main

import (
	"testing"

	"github.com/ory/kratos/examples/go/pkg"

	"github.com/stretchr/testify/require"
)

func TestSettings(t *testing.T) {
	publicURL, _ := pkg.NewKratosServer(t)
	client = pkg.NewSDKForSelfHosted(publicURL)

	flow := initSettings()
	require.NotEmpty(t, flow.Id)
}
