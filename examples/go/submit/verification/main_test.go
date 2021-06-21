package main

import (
	"testing"

	"github.com/ory/kratos/examples/go/pkg"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

func TestVerification(t *testing.T) {
	publicURL, _ := pkg.NewKratosServer(t)
	client = pkg.NewSDKForSelfHosted(publicURL)

	flow := performVerification("dev+" + uuid.New().String() + "@ory.sh")
	require.NotEmpty(t, flow.Id)
}
