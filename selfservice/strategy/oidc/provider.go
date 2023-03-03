// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/url"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"golang.org/x/oauth2"

	"github.com/ory/kratos/x"
)

type Provider interface {
	Config() *Configuration
	OAuth2(ctx context.Context) (*oauth2.Config, error)
	Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error)
	AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption
}

type TokenExchanger interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

// ConvertibleBoolean is used as Apple casually sends the email_verified field as a string.
type Claims struct {
	Issuer              string                 `json:"iss,omitempty"`
	Subject             string                 `json:"sub,omitempty"`
	Name                string                 `json:"name,omitempty"`
	GivenName           string                 `json:"given_name,omitempty"`
	FamilyName          string                 `json:"family_name,omitempty"`
	LastName            string                 `json:"last_name,omitempty"`
	MiddleName          string                 `json:"middle_name,omitempty"`
	Nickname            string                 `json:"nickname,omitempty"`
	PreferredUsername   string                 `json:"preferred_username,omitempty"`
	Profile             string                 `json:"profile,omitempty"`
	Picture             string                 `json:"picture,omitempty"`
	Website             string                 `json:"website,omitempty"`
	Email               string                 `json:"email,omitempty"`
	EmailVerified       x.ConvertibleBoolean   `json:"email_verified,omitempty"`
	Gender              string                 `json:"gender,omitempty"`
	Birthdate           string                 `json:"birthdate,omitempty"`
	Zoneinfo            string                 `json:"zoneinfo,omitempty"`
	Locale              string                 `json:"locale,omitempty"`
	PhoneNumber         string                 `json:"phone_number,omitempty"`
	PhoneNumberVerified bool                   `json:"phone_number_verified,omitempty"`
	UpdatedAt           int64                  `json:"updated_at,omitempty"`
	HD                  string                 `json:"hd,omitempty"`
	Team                string                 `json:"team,omitempty"`
	RawClaims           map[string]interface{} `json:"raw_claims,omitempty"`
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

func UpstreamParameters(provider Provider, sentProviders map[string]string) ([]oauth2.AuthCodeOption, error) {
	// https://developers.google.com/identity/openid-connect/openid-connect#authenticationuriparameters
	// Kratos already sets some parameters and we don't want to override them.
	// We also don't want to allow arbitrary parameters to be set since this could be used to craft URLs with
	// unexpected behavior.
	allowedParameters := map[string]struct{}{
		// `login_hint` sets the email address or google account id `sub` that should be pre-selected in the Google login page.
		"login_hint": {},
		// `hd` sets the organisation that is pre-selected in the Google login page.
		// Note this is a client-side setting and thus cannot be relied on to "force" a user to login
		// with a specific organisation.
		"hd": {},
	}

	params := make([]oauth2.AuthCodeOption, len(sentProviders))
	for up, v := range sentProviders {
		if _, ok := allowedParameters[up]; !ok {
			return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("The parameter %s is not allowed.", up))
		}
		params = append(params, oauth2.SetAuthURLParam(up, v))
	}

	return params, nil
}
