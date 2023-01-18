// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/oidc"
)

const wellknownJWKs = "https://raw.githubusercontent.com/aeneasr/private-oidc/master/jwks"
const wellknownToken = "https://raw.githubusercontent.com/aeneasr/private-oidc/master/token"
const fakeJWTJWKS = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjk5OTk5OTk5OTksImF1ZCI6ImFiY2QiLCJpc3MiOiJodHRwczovL3Jhdy5naXRodWJ1c2VyY29udGVudC5jb20vYWVuZWFzci9wcml2YXRlLW9pZGMvbWFzdGVyL2p3a3MifQ.RLR3dSRGGIjbRqyOMGMYFGTzcVHi7hPuFs_IKYywVWJ_XMyzWozTW4M8uuvBUPiVoNDNs7osm-AkRl7cBfw0by1XEcnEKZStCjdEh7Q0IGGb4hgq8rRqm1d3uJwNIGU5h7-s7tMnDED2ZTZhp304U99YWz7Ozl_TA9tqolBLLZEmIfXSY_RR3rMoDwtHZvWhI0OZtPdcBh86vWS9zG6QPHM5qGtRMMIs-ljXrrgS8LulUI5CAVEeHlQLXroBIe9v89IkKi07A7YRrk1SxFxlojcZ2v0z-0iTI3WL8mUoocF-RYy1RgJTK_dPYkSJebaN0R5MmBax5MXLKy4baNHKsg"
const fakeJWTToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjk5OTk5OTk5OTksImF1ZCI6ImFiY2QiLCJpc3MiOiJodHRwczovL3Jhdy5naXRodWJ1c2VyY29udGVudC5jb20vYWVuZWFzci9wcml2YXRlLW9pZGMvbWFzdGVyL3Rva2VuIn0.G9v8pJXJrEOgdJ5ecE6sIIcTH_p-RKkBaImfZY5DDVCl7h5GEis1n3GKKYbL_O3fj8Fu-WzI2mquI8S8BOVCQ6wN0XtrqJv22iX_nzeVHc4V_JWV1q7hg2gPpoFFcnF3KKtxZLvDOA8ujsDbAXmoBu0fEBdwCN56xLOOKQDzULyfijuAa8hrCwespZ9HaqcHzD3iHf_Utd4nHqlTM-6upWpKIMkplS_NGcxrfIRIWusZ0wob6ryy8jECD9QeZpdTGUozq-YM64lZfMOZzuLuqichH_PCMKFyB_tOZb6lDIiiSX4Irz7_YF-DP-LmfxgIW4934RqTCeFGGIP64h4xAA"

func TestProviderPrivateIP(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeyClientHTTPNoPrivateIPRanges, true)

	generic := func(c *oidc.Configuration) oidc.Provider {
		return oidc.NewProviderGenericOIDC(c, reg)
	}
	auth0 := func(c *oidc.Configuration) oidc.Provider {
		return oidc.NewProviderAuth0(c, reg)
	}
	gitlab := func(c *oidc.Configuration) oidc.Provider {
		return oidc.NewProviderGitLab(c, reg)
	}

	// We do not test the Auth URL as the Auth URL is not vulnerable to SSRF attacks.
	// The AuthURL is only given to the user's browser, thus it is not possible to cause SSRF.
	// We only care about Token URLs and Issuer URLs.
	for k, tc := range []struct {
		c  *oidc.Configuration
		e  string
		p  func(c *oidc.Configuration) oidc.Provider
		id string
	}{
		// Apple uses a fixed token URL and does not use the issuer.

		{p: auth0, c: &oidc.Configuration{IssuerURL: "http://127.0.0.2/"}, e: "127.0.0.2 is not a public IP address"},
		// The TokenURL is fixed in Auth0 to {issuer_url}/token. Since the issuer is called first, any local token fails also.

		// If the issuer URL is local, we fail
		{p: generic, c: &oidc.Configuration{IssuerURL: "http://127.0.0.2/"}, e: "127.0.0.2 is not a public IP address", id: fakeJWTJWKS},
		// If the issuer URL has a local JWKs URL, we fail
		{p: generic, c: &oidc.Configuration{ClientID: "abcd", IssuerURL: wellknownJWKs}, e: "is not a public IP address", id: fakeJWTJWKS},
		// The next call does not fail because the provider uses only the ID JSON Web Token to verify this call and does
		// not use the TokenURL at all!
		// {p: generic, c: &oidc.Configuration{ClientID: "abcd", IssuerURL: wellknownToken, TokenURL: "http://127.0.0.3/"}, e: "127.0.0.3 is not a public IP address", id: fakeJWTToken},

		// Discord uses a fixed token URL and does not use the issuer.
		// Facebook uses a fixed token URL and does not use the issuer.
		// GitHub uses a fixed token URL and does not use the issuer.
		// GitHub App uses a fixed token URL and does not use the issuer.
		// GitHub App uses a fixed token URL and does not use the issuer.

		{p: gitlab, c: &oidc.Configuration{IssuerURL: "http://127.0.0.2/"}, e: "127.0.0.2 is not a public IP address"},
		// The TokenURL is fixed in GitLab to {issuer_url}/token. Since the issuer is called first, any local token fails also.

		// Google uses a fixed token URL and does not use the issuer.
		// Microsoft uses a fixed token URL and does not use the issuer.
		// Slack uses a fixed token URL and does not use the issuer.
		// Spotify uses a fixed token URL and does not use the issuer.
		// VK uses a fixed token URL and does not use the issuer.
		// Yandex uses a fixed token URL and does not use the issuer.
		// NetID uses a fixed token URL and does not use the issuer.
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			p := tc.p(tc.c)
			_, err := p.Claims(context.Background(), (&oauth2.Token{RefreshToken: "foo", Expiry: time.Now().Add(-time.Hour)}).WithExtra(map[string]interface{}{
				"id_token": tc.id,
			}), url.Values{})
			require.Error(t, err)
			assert.Contains(t, fmt.Sprintf("%+v", err), tc.e)
		})
	}
}
