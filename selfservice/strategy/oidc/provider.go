// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dghubble/oauth1"

	"github.com/ory/kratos/selfservice/strategy/oidc/claims"

	"golang.org/x/oauth2"
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
// - `login_hint` (string): The `login_hint` parameter suppresses the account chooser and either pre-fills the email box on the sign-in form, or selects the proper session.
// - `hd` (string): The `hd` parameter limits the login/registration process to a Google Organization, e.g. `mycollege.edu`.
// - `prompt` (string): The `prompt` specifies whether the Authorization Server prompts the End-User for reauthentication and consent, e.g. `select_account`.
// - `auth_type` (string): The `auth_type` parameter specifies the requested authentication features (as a comma-separated list), e.g. `reauthenticate`.
func UpstreamParameters(upstreamParameters map[string]string) []oauth2.AuthCodeOption {
	// validation of upstream parameters are already handled in the `oidc/.schema/link.schema.json` and `oidc/.schema/settings.schema.json` file.
	// `upstreamParameters` will always only contain allowed parameters based on the configuration.

	// we double-check the parameters here to prevent any potential security issues.
	allowedParameters := map[string]struct{}{
		"login_hint": {},
		"hd":         {},
		"prompt":     {},
		"auth_type":  {},
	}

	var params []oauth2.AuthCodeOption
	for up, v := range upstreamParameters {
		if _, ok := allowedParameters[up]; ok {
			params = append(params, oauth2.SetAuthURLParam(up, v))
		}
	}

	return params
}
