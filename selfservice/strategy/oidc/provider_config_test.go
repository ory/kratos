package oidc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestConfig(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	var c map[string]interface{}
	require.NoError(t, json.NewDecoder(
		bytes.NewBufferString(`{"config":{"providers": [{"provider": "generic"}]}}`)).Decode(&c))
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC), c)

	s := oidc.NewStrategy(reg)
	collection, err := s.Config(context.Background())
	require.NoError(t, err)

	require.Len(t, collection.Providers, 1)
	assert.Equal(t, "generic", collection.Providers[0].Provider)
}

func TestConfig_GetCustomProvider(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	oidc.CustomProviders["custom"] = func(config *oidc.Configuration, public *url.URL) oidc.Provider {
		return &TestProvider{}
	}
	defer func() {
		delete(oidc.CustomProviders, "custom")
	}()

	var c map[string]interface{}
	require.NoError(t, json.NewDecoder(
		bytes.NewBufferString(`{"config":{"providers": [{"id": "custom", "provider": "custom"}]}}`)).Decode(&c))
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC), c)

	s := oidc.NewStrategy(reg)
	collection, err := s.Config(context.Background())
	require.NoError(t, err)

	pubURL, err := url.Parse("http://127.0.0.1/foo")
	require.NoError(t, err)

	p, err := collection.Provider("custom", pubURL)
	require.NoError(t, err)
	require.IsType(t, &TestProvider{}, p)
}

func TestConfig_GetCustomProvider_NotRegistered(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	var c map[string]interface{}
	require.NoError(t, json.NewDecoder(
		bytes.NewBufferString(`{"config":{"providers": [{"id": "custom", "provider": "custom"}]}}`)).Decode(&c))
	conf.MustSet(config.ViperKeySelfServiceStrategyConfig+"."+string(identity.CredentialsTypeOIDC), c)

	s := oidc.NewStrategy(reg)
	collection, err := s.Config(context.Background())
	require.NoError(t, err)

	pubURL, err := url.Parse("http://127.0.0.1/foo")
	require.NoError(t, err)

	_, err = collection.Provider("custom", pubURL)
	require.Error(t, err)
}

type TestProvider struct{}

func (t TestProvider) Config() *oidc.Configuration {
	panic("not called")
}

func (t TestProvider) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	panic("not called")
}

func (t TestProvider) Claims(ctx context.Context, exchange *oauth2.Token) (*oidc.Claims, error) {
	panic("not called")
}

func (t TestProvider) AuthCodeURLOptions(r oidc.Ider) []oauth2.AuthCodeOption {
	panic("not called")
}
