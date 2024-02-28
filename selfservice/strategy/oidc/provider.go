// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/http"
	"net/url"

	"github.com/dghubble/oauth1"
	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"golang.org/x/oauth2"

	"github.com/ory/kratos/x"
)

type Provider interface {
	Config() *Configuration
}

type OAuth2Provider interface {
	Provider
	AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption
	OAuth2(ctx context.Context) (*oauth2.Config, error)
	Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error)
}

type OAuth1Provider interface {
	Provider
	OAuth1(ctx context.Context) *oauth1.Config
	AuthURL(ctx context.Context, state string) (string, error)
	Claims(ctx context.Context, token *oauth1.Token) (*Claims, error)
	ExchangeToken(ctx context.Context, req *http.Request) (*oauth1.Token, error)
}

type OAuth2TokenExchanger interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type IDTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*Claims, error)
}

type NonceValidationSkipper interface {
	CanSkipNonce(*Claims) bool
}

// ConvertibleBoolean is used as Apple casually sends the email_verified field as a string.
type Claims struct {
	Issuer              string                 `json:"iss"`
	Subject             string                 `json:"sub"`
	Name                string                 `json:"name"`
	GivenName           string                 `json:"given_name"`
	FamilyName          string                 `json:"family_name"`
	LastName            string                 `json:"last_name"`
	MiddleName          string                 `json:"middle_name"`
	Nickname            string                 `json:"nickname"`
	PreferredUsername   string                 `json:"preferred_username"`
	Profile             string                 `json:"profile"`
	Picture             string                 `json:"picture"`
	Website             string                 `json:"website"`
	Email               string                 `json:"email"`
	EmailVerified       x.ConvertibleBoolean   `json:"email_verified"`
	Gender              string                 `json:"gender"`
	Birthdate           string                 `json:"birthdate"`
	Zoneinfo            string                 `json:"zoneinfo"`
	Locale              string                 `json:"locale"`
	PhoneNumber         string                 `json:"phone_number"`
	PhoneNumberVerified bool                   `json:"phone_number_verified"`
	UpdatedAt           int64                  `json:"updated_at"`
	HD                  string                 `json:"hd"`
	Team                string                 `json:"team"`
	Nonce               string                 `json:"nonce"`
	NonceSupported      bool                   `json:"nonce_supported"`
	RawClaims           map[string]interface{} `json:"raw_claims"`
}

// Validate checks if the claims are valid.
func (c *Claims) Validate() error {
	if c.Subject == "" {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("provider did not return a subject"))
	}
	if c.Issuer == "" {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("issuer not set in claims"))
	}
	return nil
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
