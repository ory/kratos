package oidc_test

import (
	"context"
	"encoding/json"
	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

func makeOIDCClaims() json.RawMessage {
	claims, _ := json.Marshal(map[string]interface{}{
		"id_token": map[string]interface{}{
			"email": map[string]bool{
				"essential": true,
			},
			"email_verified": map[string]bool{
				"essential": true,
			},
		},
	})
	return claims
}

func makeAuthCodeURL(t *testing.T, r *login.Flow, reg *driver.RegistryDefault) string {
	p := oidc.NewProviderGenericOIDC(&oidc.Configuration{
		Provider:        "generic",
		ID:              "valid",
		ClientID:        "client",
		ClientSecret:    "secret",
		IssuerURL:       "https://accounts.google.com",
		Mapper:          "file://./stub/hydra.schema.json",
		RequestedClaims: makeOIDCClaims(),
	}, reg)
	c, err := p.OAuth2(context.Background())
	require.NoError(t, err)
	return c.AuthCodeURL("state", p.AuthCodeURLOptions(r)...)
}

func TestProviderGenericOIDC_AddAuthCodeURLOptions(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(config.ViperKeyPublicBaseURL, "https://ory.sh")
	t.Run("case=expect prompt to be login with forced flag", func(t *testing.T) {
		r := &login.Flow{
			ID:      x.NewUUID(),
			Refresh: true,
		}
		assert.Contains(t, makeAuthCodeURL(t, r, reg), "prompt=login")
	})

	t.Run("case=expect prompt to not be login without forced flag", func(t *testing.T) {
		r := &login.Flow{
			ID: x.NewUUID(),
		}
		assert.NotContains(t, makeAuthCodeURL(t, r, reg), "prompt=login")
	})

	t.Run("case=expect requested claims to be set", func(t *testing.T) {
		r := &login.Flow{
			ID: x.NewUUID(),
		}
		assert.Contains(t, makeAuthCodeURL(t, r, reg), "claims="+url.QueryEscape(string(makeOIDCClaims())))
	})

	t.Run("case=does not allow calling issuers in the internal network", func(t *testing.T) {

	})

	t.Run("case=does not allow calling mappers in the internal network", func(t *testing.T) {

	})
}
