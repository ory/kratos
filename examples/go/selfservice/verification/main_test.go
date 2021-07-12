package main

import (
	"testing"

	ory "github.com/ory/kratos-client-go"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/kratos/examples/go/pkg"

	"github.com/stretchr/testify/require"
)

func TestFunc(t *testing.T) {
	publicURL, _ := testhelpers.StartE2EServer(t, "../../pkg/stub/kratos.yaml", nil)
	client = pkg.NewSDKForSelfHosted(publicURL)

	flow := performVerification("dev+" + uuid.New().String() + "@ory.sh")
	require.NotEmpty(t, flow.Id)
	assert.Equal(t, ory.SELFSERVICEVERIFICATIONFLOWSTATE_SENT_EMAIL, flow.State)
}
