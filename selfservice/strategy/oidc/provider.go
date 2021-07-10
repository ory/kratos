package oidc

import (
	"bytes"
	"context"
	"encoding/json"

	"golang.org/x/oauth2"
)

type Provider interface {
	Config() *Configuration
	OAuth2(ctx context.Context) (*oauth2.Config, error)
	Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error)
	AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption
}

func toClaims(basicClaims BasicClaims) Claims {
	var buffer []byte
	bb := bytes.NewBuffer(buffer)
	enc := json.NewEncoder(bb)
	err := enc.Encode(basicClaims)
	if err != nil {
		return Claims{}
	}

	dec := json.NewDecoder(bb)
	var claims Claims
	err = dec.Decode(&claims)
	if err != nil {
		return Claims{}
	}

	return claims
}

type Claims struct {
	BasicClaims
	AdditionalClaims map[string]interface{} `json:"additional_claims"`
}

type BasicClaims struct {
	Issuer              string `json:"iss,omitempty"`
	Subject             string `json:"sub,omitempty"`
	Name                string `json:"name,omitempty"`
	GivenName           string `json:"given_name,omitempty"`
	FamilyName          string `json:"family_name,omitempty"`
	LastName            string `json:"last_name,omitempty"`
	MiddleName          string `json:"middle_name,omitempty"`
	Nickname            string `json:"nickname,omitempty"`
	PreferredUsername   string `json:"preferred_username,omitempty"`
	Profile             string `json:"profile,omitempty"`
	Picture             string `json:"picture,omitempty"`
	Website             string `json:"website,omitempty"`
	Email               string `json:"email,omitempty"`
	EmailVerified       bool   `json:"email_verified,omitempty"`
	Gender              string `json:"gender,omitempty"`
	Birthdate           string `json:"birthdate,omitempty"`
	Zoneinfo            string `json:"zoneinfo,omitempty"`
	Locale              string `json:"locale,omitempty"`
	PhoneNumber         string `json:"phone_number,omitempty"`
	PhoneNumberVerified bool   `json:"phone_number_verified,omitempty"`
	UpdatedAt           int64  `json:"updated_at,omitempty"`
	HD                  string `json:"hd,omitempty"`
	Team                string `json:"team,omitempty"`
}
