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

	"github.com/golang-jwt/jwt/v4"
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

// Just a public key set, to be able to test what happens if an ID token was issued by a different private key.
//
//go:embed stub/jwks_public2.json
var publicJWKS2 []byte

type claims struct {
	*jwt.RegisteredClaims
	Email string `json:"email"`
}

func createIdToken(t *testing.T, aud string) string {
	key := &jwk.KeySpec{}
	require.NoError(t, json.Unmarshal(rawKey, key))
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &claims{
		RegisteredClaims: &jwt.RegisteredClaims{
			Issuer:    "https://appleid.apple.com",
			Subject:   "apple@ory.sh",
			Audience:  jwt.ClaimStrings{aud},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
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

	tsOtherJWKS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(publicJWKS2)
	}))
	t.Run("case=successful verification", func(t *testing.T) {
		apple := ProviderApple{
			jwksUrl: ts.URL,
			ProviderGenericOIDC: &ProviderGenericOIDC{
				config: &Configuration{
					ClientID: "com.example.app",
				},
			},
		}
		token := createIdToken(t, "com.example.app")

		c, err := apple.Verify(context.Background(), token)
		require.NoError(t, err)
		assert.Equal(t, "apple@ory.sh", c.Email)
		assert.Equal(t, "apple@ory.sh", c.Subject)
		assert.Equal(t, "https://appleid.apple.com", c.Issuer)
	})

	t.Run("case=fails due to client_id mismatch", func(t *testing.T) {
		apple := ProviderApple{
			jwksUrl: ts.URL,
			ProviderGenericOIDC: &ProviderGenericOIDC{
				config: &Configuration{
					ClientID: "com.example.app",
				},
			},
		}
		token := createIdToken(t, "com.different-example.app")

		_, err := apple.Verify(context.Background(), token)
		require.Error(t, err)
		assert.Equal(t, `oidc: expected audience "com.example.app" got ["com.different-example.app"]`, err.Error())
	})

	t.Run("case=fails due to jwks mismatch", func(t *testing.T) {
		apple := ProviderApple{
			jwksUrl: tsOtherJWKS.URL,
			ProviderGenericOIDC: &ProviderGenericOIDC{
				config: &Configuration{
					ClientID: "com.example.app",
				},
			},
		}
		token := createIdToken(t, "com.example.app")

		_, err := apple.Verify(context.Background(), token)
		require.Error(t, err)
		assert.Equal(t, "failed to verify signature: failed to verify id token signature", err.Error())
	})
}
