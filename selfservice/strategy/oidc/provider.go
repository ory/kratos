// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dghubble/oauth1"
	"golang.org/x/oauth2"

	"github.com/ory/kratos/selfservice/strategy/oidc/claims"
)

type (
	Provider interface {
		Config() *Configuration
	}
	OAuth2Provider interface {
		Provider
		AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption
		OAuth2(ctx context.Context) (*oauth2.Config, error)
		Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*claims.Claims, error)
	}
	OAuth1Provider interface {
		Provider
		OAuth1(ctx context.Context) *oauth1.Config
		AuthURL(ctx context.Context, state string) (string, error)
		Claims(ctx context.Context, token *oauth1.Token) (*claims.Claims, error)
		ExchangeToken(ctx context.Context, req *http.Request) (*oauth1.Token, error)
	}
)

type OAuth2TokenExchanger interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type IDTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*claims.Claims, error)
}

type NonceValidationSkipper interface {
	CanSkipNonce(*claims.Claims) bool
}

// UpstreamParameters returns a list of oauth2.AuthCodeOption based on the upstream parameters.
//
// Only allowed parameters are returned and the rest is ignored.
// Allowed parameters are also defined in the `oidc/.schema/link.schema.json` file, however,
// this function also validates the parameters to prevent any potential security issues.
//
// Allowed parameters are:
// - `hd` (string): The `hd` parameter limits the login/registration process to a Google Organization, e.g. `mycollege.edu`.
// - `prompt` (string): The `prompt` specifies whether the Authorization Server prompts the End-User for reauthentication and consent, e.g. `select_account`.
// - `auth_type` (string): The `auth_type` parameter specifies the requested authentication features (as a comma-separated list), e.g. `reauthenticate`.
// - `acr_values` (string): The `acr_values` specifies the Authentication Context Class Reference values for the authorization request.
//
// `login_hint` is intentionally not forwarded. A caller-supplied value cannot be
// trusted, and on IdPs that act on it (Google, Microsoft, Auth0, …) it can steer
// account, tenant, or OAuth client selection and enable a client-swap attack
// (HackerOne #3239672, ory-corp/cloud#8955).
func UpstreamParameters(upstreamParameters map[string]string) []oauth2.AuthCodeOption {
	// validation of upstream parameters are already handled in the `oidc/.schema/link.schema.json` and `oidc/.schema/settings.schema.json` file.
	// `upstreamParameters` will always only contain allowed parameters based on the configuration.

	// we double-check the parameters here to prevent any potential security issues.
	allowedParameters := map[string]struct{}{
		"hd":         {},
		"prompt":     {},
		"auth_type":  {},
		"acr_values": {},
	}

	var params []oauth2.AuthCodeOption
	for up, v := range upstreamParameters {
		if _, ok := allowedParameters[up]; !ok {
			continue
		}
		params = append(params, oauth2.SetAuthURLParam(up, v))
	}

	return params
}
