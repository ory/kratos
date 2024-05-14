// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package claims

import (
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/x"
)

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
