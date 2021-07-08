package main

import (
	"testing"

	ory "github.com/ory/kratos-client-go"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/kratos/examples/go/pkg"

	"github.com/stretchr/testify/require"
)

func TestSettings(t *testing.T) {
	publicURL, _ := testhelpers.StartE2EServer(t, "../../pkg/stub/kratos.yaml")
	client = pkg.NewSDKForSelfHosted(publicURL)

	email, password := pkg.RandomCredentials()
	result := changePassword(email, password)
	require.NotEmpty(t, result.Flow.Id)
	assert.Equal(t, ory.SELFSERVICESETTINGSFLOWSTATE_SUCCESS, result.Flow.State)

	email, password = pkg.RandomCredentials()
	result = changeTraits(email, password)
	require.NotEmpty(t, result.Flow.Id)
	assert.Equal(t, ory.SELFSERVICESETTINGSFLOWSTATE_SUCCESS, result.Flow.State)
	assert.Equal(t, "not-"+email, result.Identity.Traits.(map[string]interface{})["email"].(string))
}
