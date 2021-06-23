package main

import (
	"testing"

	"github.com/ory/kratos/examples/go/pkg"

	"github.com/stretchr/testify/require"
)

func TestRegistration(t *testing.T) {
	publicURL, _ := pkg.NewKratosServer(t)
	client = pkg.NewSDKForSelfHosted(publicURL)

	flow := initRegistration()
	require.NotEmpty(t, flow.Id)
}
