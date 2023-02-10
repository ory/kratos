package oidc_test

import (
	"context"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/x"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"testing"
)

func TestProviderGoogle_Scope(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)

	p := oidc.NewProviderGoogle(&oidc.Configuration{
		Provider:        "google",
		ID:              "valid",
		ClientID:        "client",
		ClientSecret:    "secret",
		Mapper:          "file://./stub/hydra.schema.json",
		RequestedClaims: nil,
		Scope:           []string{"email", "profile", "offline_access"},
	}, reg)

	c, _ := p.OAuth2(context.Background())
	assert.NotContains(t, c.Scopes, "offline_access")
}

func TestProviderGoogle_AccessType(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)

	p := oidc.NewProviderGoogle(&oidc.Configuration{
		Provider:        "google",
		ID:              "valid",
		ClientID:        "client",
		ClientSecret:    "secret",
		Mapper:          "file://./stub/hydra.schema.json",
		RequestedClaims: nil,
		Scope:           []string{"email", "profile", "offline_access"},
	}, reg)

	r := &login.Flow{
		ID: x.NewUUID(),
	}

	options := p.AuthCodeURLOptions(r)
	assert.Contains(t, options, oauth2.AccessTypeOffline)
}
