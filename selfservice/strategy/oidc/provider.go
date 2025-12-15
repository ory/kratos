// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/dghubble/oauth1"
	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"golang.org/x/oauth2"

	"github.com/ory/kratos/x"
)

type (
	Provider interface {
		Config() *Configuration
	}
	OAuth2Provider interface {
		Provider
		AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption
		OAuth2(ctx context.Context) (*oauth2.Config, error)
		Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error)
	}
	OAuth1Provider interface {
		Provider
		OAuth1(ctx context.Context) *oauth1.Config
		AuthURL(ctx context.Context, state string) (string, error)
		Claims(ctx context.Context, token *oauth1.Token) (*Claims, error)
		ExchangeToken(ctx context.Context, req *http.Request) (*oauth1.Token, error)
	}
)

type OAuth2TokenExchanger interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type IDTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*Claims, error)
}

type NonceValidationSkipper interface {
	CanSkipNonce(*Claims) bool
}

type Claims struct {
	Issuer            string `json:"iss,omitempty"`
	Subject           string `json:"sub,omitempty"`
	Object            string `json:"oid,omitempty"`
	Name              string `json:"name,omitempty"`
	GivenName         string `json:"given_name,omitempty"`
	FamilyName        string `json:"family_name,omitempty"`
	LastName          string `json:"last_name,omitempty"`
	MiddleName        string `json:"middle_name,omitempty"`
	Nickname          string `json:"nickname,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Profile           string `json:"profile,omitempty"`
	Picture           string `json:"picture,omitempty"`
	Website           string `json:"website,omitempty"`
	Email             string `json:"email,omitempty"`
	// ConvertibleBoolean is used as Apple casually sends the email_verified field as a string.
	EmailVerified       x.ConvertibleBoolean   `json:"email_verified,omitempty"`
	Gender              string                 `json:"gender,omitempty"`
	Birthdate           string                 `json:"birthdate,omitempty"`
	Zoneinfo            string                 `json:"zoneinfo,omitempty"`
	Locale              Locale                 `json:"locale,omitempty"`
	PhoneNumber         string                 `json:"phone_number,omitempty"`
	PhoneNumberVerified bool                   `json:"phone_number_verified,omitempty"`
	UpdatedAt           int64                  `json:"updated_at,omitempty"`
	HD                  string                 `json:"hd,omitempty"`
	Team                string                 `json:"team,omitempty"`
	Nonce               string                 `json:"nonce,omitempty"`
	NonceSupported      bool                   `json:"nonce_supported,omitempty"`
	RawClaims           map[string]interface{} `json:"raw_claims,omitempty"`
}

type Locale string

func (l *Locale) UnmarshalJSON(data []byte) error {
	var linkedInLocale struct {
		Language string `json:"language"`
		Country  string `json:"country"`
	}
	if err := json.Unmarshal(data, &linkedInLocale); err == nil {
		switch {
		case linkedInLocale.Language == "":
			*l = Locale(linkedInLocale.Country)
		case linkedInLocale.Country == "":
			*l = Locale(linkedInLocale.Language)
		default:
			*l = Locale(strings.Join([]string{linkedInLocale.Language, linkedInLocale.Country}, "-"))
		}

		return nil
	}

	return json.Unmarshal(data, (*string)(l))
}

// Validate checks if the claims are valid.
func (c *Claims) Validate() error {
	if c.Subject == "" {
		return errors.WithStack(herodot.ErrUpstreamError.WithReasonf("provider did not return a subject"))
	}
	if c.Issuer == "" {
		return errors.WithStack(herodot.ErrUpstreamError.WithReasonf("issuer not set in claims"))
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
