// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	_ "embed"

	"github.com/form3tech-oss/jwt-go"
	"github.com/rakutentech/jwk-go/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeQuery(t *testing.T) {
	query := url.Values{
		"user": []string{`{"name": {"firstName": "first", "lastName": "last"}, "email": "email@email.com"}`},
	}

	for k, tc := range []struct {
		claims     *Claims
		familyName string
		givenName  string
		lastName   string
	}{
		{claims: &Claims{}, familyName: "first", givenName: "first", lastName: "last"},
		{claims: &Claims{FamilyName: "fam"}, familyName: "fam", givenName: "first", lastName: "last"},
		{claims: &Claims{FamilyName: "fam", GivenName: "giv"}, familyName: "fam", givenName: "giv", lastName: "last"},
		{claims: &Claims{FamilyName: "fam", GivenName: "giv", LastName: "las"}, familyName: "fam", givenName: "giv", lastName: "las"},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			decodeQuery(query, tc.claims)
			assert.Equal(t, tc.familyName, tc.claims.FamilyName)
			assert.Equal(t, tc.givenName, tc.claims.GivenName)
			assert.Equal(t, tc.lastName, tc.claims.LastName)
			// Never extract email from the query, as the same info can be extracted and verified from the ID token.
			assert.Empty(t, tc.claims.Email)
		})
	}

}

//go:embed stub/jwk.json
var rawKey []byte

//go:embed stub/jwks_public.json
var publicJWKS []byte

type claims struct {
	*jwt.StandardClaims
	Email string `json:"email"`
}

func createIdToken(t *testing.T) string {
	key := &jwk.KeySpec{}
	require.NoError(t, json.Unmarshal(rawKey, key))
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &claims{
		StandardClaims: &jwt.StandardClaims{
			Issuer:    "https://appleid.apple.com",
			Subject:   "apple@ory.sh",
			Audience:  []string{"com.example.app"},
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
		Email: "apple@ory.sh",
	})
	token.Header["kid"] = key.KeyID
	s, err := token.SignedString(key.Key)
	require.NoError(t, err)
	return s
}

func TestVerify(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(publicJWKS)
	}))
	apple := ProviderApple{
		jwksUrl: ts.URL,
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: &Configuration{
				ClientID: "com.example.app",
			},
		},
	}
	token := createIdToken(t)

	c, err := apple.Verify(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, "apple@ory.sh", c.Email)
	assert.Equal(t, "apple@ory.sh", c.Subject)
	assert.Equal(t, "https://appleid.apple.com", c.Issuer)

}
