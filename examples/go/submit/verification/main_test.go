package main

import (
	"testing"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/kratos/examples/go/pkg"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

func TestVerification(t *testing.T) {
	publicURL, _ := testhelpers.StartE2EServer(t, "../../pkg/stub/kratos.yaml")
	client = pkg.NewSDKForSelfHosted(publicURL)

	flow := performVerification("dev+" + uuid.New().String() + "@ory.sh")
	require.NotEmpty(t, flow.Id)
}
