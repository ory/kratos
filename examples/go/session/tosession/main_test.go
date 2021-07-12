package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/kratos/examples/go/pkg"

	"github.com/stretchr/testify/require"
)

func TestFunc(t *testing.T) {
	publicURL, _ := testhelpers.StartE2EServer(t, "../../pkg/stub/kratos.yaml")
	client = pkg.NewSDKForSelfHosted(publicURL)

	session := toSession()
	require.NotEmpty(t, session.Id)
	assert.Contains(t, session.Identity.Traits.(map[string]interface{})["email"].(string), "dev+")
}
