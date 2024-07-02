// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

type User struct {
	SerialNumber      string
	ParseSerialNumber (oidc.ProviderEParaksts)
}

func TestProviderEparaksts_ParseSerialNumber(t *testing.T) {
	expected := "1234678900"
	serialNumber := "PVOLV-12346-78900"

	g := &oidc.ProviderEParaksts{}

	actual := g.ParseSerialNumber(serialNumber)
	assert.Equal(t, expected, actual)
}

func TestProviderEparaksts_Discovery(t *testing.T) {
	_, reg := internal.NewVeryFastRegistryWithoutDB(t)

	p := oidc.NewProviderEParakstsMobile(&oidc.Configuration{
		ID:           "eparaksts",
		Provider:     "eparaksts",
		ClientID:     "abcd",
		ClientSecret: "secret",
		IssuerURL:    "https://eidas.eparaksts.lv",
		Mapper:       "file://./stub/oidc.eparaksts.jsonnet",
		Scope:        []string{"urn:lvrtc:fpeil:aa"},
	}, reg)

	c, err := p.(oidc.OAuth2Provider).OAuth2(context.Background())
	require.NoError(t, err)
	assert.Contains(t, c.Scopes, "urn:lvrtc:fpeil:aa")
	assert.Equal(t, "https://eidas.eparaksts.lv/trustedx-authserver/oauth/lvrtc-eipsign-as/token", c.Endpoint.TokenURL)
	assert.Equal(t, "https://eidas.eparaksts.lv/trustedx-authserver/oauth/lvrtc-eipsign-as?acr_values=urn%3Aeparaksts%3Aauthentication%3Aflow%3Amobileid&prompt=login&ui_locales=lv", c.Endpoint.AuthURL)
}

func TestProviderEparaksts_Claims(t *testing.T) {
	_, base := internal.NewFastRegistryWithMocks(t)
	reg := &mockRegistry{base, retryablehttp.NewClient()}
	httpmock.ActivateNonDefault(reg.cl.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)

	httpmock.RegisterResponder("GET", "https://eidas-demo.eparaksts.lv/trustedx-resources/openid/v1/users/me",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"sub":           "5foOWOiYXD",
				"acr":           "urn:eparaksts:authentication:flow:sc_plugin",
				"domain":        "citizen",
				"given_name":    "John",
				"family_name":   "Doe",
				"name":          "John Doe",
				"serial_number": "PVOLV-12346-78900",
			})
			return resp, err
		})

	c := &oidc.Configuration{
		ID:           "eparaksts",
		Provider:     "eparaksts",
		ClientID:     "abcd",
		ClientSecret: "secret",
		IssuerURL:    "https://eidas-demo.eparaksts.lv",
		Mapper:       "file://./stub/oidc.eparaksts.jsonnet",
		Scope:        []string{"urn:lvrtc:fpeil:aa"},
	}
	eparaksts := oidc.NewProviderEParaksts(c, reg)

	actual, err := eparaksts.(oidc.OAuth2Provider).Claims(
		context.Background(),
		(&oauth2.Token{AccessToken: "foo", Expiry: time.Now().Add(time.Hour)}),
		url.Values{},
	)
	require.NoError(t, err)

	assert.Equal(t, &oidc.Claims{
		Issuer:    "https://eidas-demo.eparaksts.lv",
		Subject:   "5foOWOiYXD",
		GivenName: "John",
		LastName:  "Doe",
		RawClaims: map[string]interface{}{
			"serial_number": "1234678900",
		},
	}, actual)
}
