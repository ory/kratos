// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/ory/kratos/x"
)

// CredentialsOIDC is contains the configuration for credentials of the type oidc.
//
// swagger:model identityCredentialsOidc
type CredentialsOIDC struct {
	Providers []CredentialsOIDCProvider `json:"providers"`
}

// CredentialsOIDCProvider is contains a specific OpenID COnnect credential for a particular connection (e.g. Google).
//
// swagger:model identityCredentialsOidcProvider
type CredentialsOIDCProvider struct {
	Subject             string `json:"subject"`
	Provider            string `json:"provider"`
	InitialIDToken      string `json:"initial_id_token"`
	InitialAccessToken  string `json:"initial_access_token"`
	InitialRefreshToken string `json:"initial_refresh_token"`
	Organization        string `json:"organization,omitempty"`
}

// swagger:ignore
type CredentialsOIDCEncryptedTokens struct {
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
}

func (c *CredentialsOIDCEncryptedTokens) GetRefreshToken() string {
	if c == nil {
		return ""
	}
	return c.RefreshToken
}

func (c *CredentialsOIDCEncryptedTokens) GetAccessToken() string {
	if c == nil {
		return ""
	}
	return c.AccessToken
}

func (c *CredentialsOIDCEncryptedTokens) GetIDToken() string {
	if c == nil {
		return ""
	}
	return c.IDToken
}

// NewCredentialsOIDC creates a new OIDC credential.
func NewCredentialsOIDC(tokens *CredentialsOIDCEncryptedTokens, provider, subject, organization string) (*Credentials, error) {
	if provider == "" {
		return nil, errors.New("received empty provider in oidc credentials")
	}

	if subject == "" {
		return nil, errors.New("received empty provider in oidc credentials")
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(CredentialsOIDC{
		Providers: []CredentialsOIDCProvider{
			{
				Subject:             subject,
				Provider:            provider,
				InitialIDToken:      tokens.GetIDToken(),
				InitialAccessToken:  tokens.GetAccessToken(),
				InitialRefreshToken: tokens.GetRefreshToken(),
				Organization:        organization,
			}},
	}); err != nil {
		return nil, errors.WithStack(x.PseudoPanic.
			WithDebugf("Unable to encode password options to JSON: %s", err))
	}

	return &Credentials{
		Type:        CredentialsTypeOIDC,
		Identifiers: []string{OIDCUniqueID(provider, subject)},
		Config:      b.Bytes(),
	}, nil
}

func (c *CredentialsOIDCProvider) GetTokens() *CredentialsOIDCEncryptedTokens {
	return &CredentialsOIDCEncryptedTokens{
		RefreshToken: c.InitialRefreshToken,
		IDToken:      c.InitialIDToken,
		AccessToken:  c.InitialAccessToken,
	}
}

func OIDCUniqueID(provider, subject string) string {
	return fmt.Sprintf("%s:%s", provider, subject)
}

func (c *CredentialsOIDC) Organization() string {
	for _, p := range c.Providers {
		if p.Organization != "" {
			return p.Organization
		}
	}

	return ""
}
