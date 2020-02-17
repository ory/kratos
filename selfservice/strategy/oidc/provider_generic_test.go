package oidc

import (
	"context"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func makeAuthCodeURL(t *testing.T, r *login.Request) string {
	public, err := url.Parse("https://ory.sh")
	require.NoError(t, err)
	p := NewProviderGenericOIDC(&Configuration{
		Provider:     "generic",
		ID:           "valid",
		ClientID:     "client",
		ClientSecret: "secret",
		IssuerURL:    "https://accounts.google.com",
		SchemaURL:    "file://./stub/hydra.schema.json",
	}, public)
	c, err := p.OAuth2(context.TODO())
	require.NoError(t, err)
	return c.AuthCodeURL("state", p.AddAuthCodeURLOptions(r)...)
}

func TestProviderGenericOIDC_AddAuthCodeURLOptions(t *testing.T) {
	t.Run("case=expect prompt to be login with reauthentication flag", func(t *testing.T) {
		r := &login.Request{
			ID: x.NewUUID(),
			IsReauthentication: true,
		}
		assert.Contains(t, makeAuthCodeURL(t, r), "prompt=login")
	})

	t.Run("case=expect prompt to not be login without reauthentication flag", func(t *testing.T) {
		r := &login.Request{
			ID: x.NewUUID(),
		}
		assert.NotContains(t, makeAuthCodeURL(t, r), "prompt=login")
	})
}
