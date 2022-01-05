package oidc

import (
	"context"
	"net/http"

	"github.com/ory/kratos/cipher"
	"github.com/ory/kratos/x"
)

// Token returned by providers
type Token interface {
	Claims(ctx context.Context) (*Claims, error)
	CredentialsConfig(ctx context.Context, c cipher.Cipher) (*ProviderCredentialsConfig, error)
}

// Provider for implementations of identity providers
type Provider interface {
	Config() *Configuration
	Token(ctx context.Context, req *http.Request) (Token, error)
	RedirectURL(ctx context.Context, state string, r ider) (string, error)
}

// ConvertibleBoolean is used as Apple casually sends the email_verified field as a string.
type Claims struct {
	Issuer              string               `json:"iss,omitempty"`
	Subject             string               `json:"sub,omitempty"`
	Name                string               `json:"name,omitempty"`
	GivenName           string               `json:"given_name,omitempty"`
	FamilyName          string               `json:"family_name,omitempty"`
	LastName            string               `json:"last_name,omitempty"`
	MiddleName          string               `json:"middle_name,omitempty"`
	Nickname            string               `json:"nickname,omitempty"`
	PreferredUsername   string               `json:"preferred_username,omitempty"`
	Profile             string               `json:"profile,omitempty"`
	Picture             string               `json:"picture,omitempty"`
	Website             string               `json:"website,omitempty"`
	Email               string               `json:"email,omitempty"`
	EmailVerified       x.ConvertibleBoolean `json:"email_verified,omitempty"`
	Gender              string               `json:"gender,omitempty"`
	Birthdate           string               `json:"birthdate,omitempty"`
	Zoneinfo            string               `json:"zoneinfo,omitempty"`
	Locale              string               `json:"locale,omitempty"`
	PhoneNumber         string               `json:"phone_number,omitempty"`
	PhoneNumberVerified bool                 `json:"phone_number_verified,omitempty"`
	UpdatedAt           int64                `json:"updated_at,omitempty"`
	HD                  string               `json:"hd,omitempty"`
	Team                string               `json:"team,omitempty"`
}
