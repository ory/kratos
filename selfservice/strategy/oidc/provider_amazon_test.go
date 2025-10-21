// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/amazon"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

func TestAmazonOidcClaims(t *testing.T) {
	t.Parallel()

	handler := http.NewServeMux()
	expectedAccessToken := "my-access-token"
	handler.HandleFunc("GET /user/profile", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("x-amz-access-token")
		if token != expectedAccessToken {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		// From the official docs: https://developer.amazon.com/docs/login-with-amazon/customer-profile.html .
		userProfile := `
{
    "user_id" : "amzn1.account.K2LI23KL2LK2",
    "email" : "johndoe@gmail.com",
    "name" : "John Doe",
    "postal_code": "98101"
}
`

		_, err := w.Write([]byte(userProfile))
		require.NoError(t, err)
	})
	amazonApi := httptest.NewServer(handler)
	t.Cleanup(amazonApi.Close)

	_, reg := internal.NewFastRegistryWithMocks(t)
	p := oidc.NewProviderAmazon(&oidc.Configuration{}, reg).(*oidc.ProviderAmazon)
	p.SetProfileURL(amazonApi.URL + "/user/profile")

	claims, err := p.Claims(t.Context(), &oauth2.Token{AccessToken: expectedAccessToken}, nil)
	require.NoError(t, err)
	require.NotNil(t, claims)

	require.Equal(t, claims.Subject, "amzn1.account.K2LI23KL2LK2")
	require.Equal(t, claims.Issuer, amazon.Endpoint.TokenURL)
	require.Equal(t, claims.Name, "John Doe")
	require.Equal(t, claims.Email, "johndoe@gmail.com")
	require.Equal(t, claims.Zoneinfo, "98101")
}
