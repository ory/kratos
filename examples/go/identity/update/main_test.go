package main

import (
	"testing"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/kratos/examples/go/pkg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunc(t *testing.T) {
	publicURL, _ := testhelpers.StartE2EServer(t, "../../pkg/stub/kratos.yaml")
	client = pkg.NewSDKForSelfHosted(publicURL)
	identity := updateIdentity()
	require.NotEmpty(t, identity.Id)
	assert.Contains(t, identity.Traits.(map[string]interface{})["email"].(string), "dev+not")
}
