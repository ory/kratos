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
	CurrentIDToken      string `json:"current_id_token"`
	CurrentAccessToken  string `json:"current_access_token"`
	CurrentRefreshToken string `json:"current_refresh_token"`
}

// NewCredentialsOIDC creates a new OIDC credential.
func NewCredentialsOIDC(idToken, accessToken, refreshToken, provider, subject string) (*Credentials, error) {
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
				InitialIDToken:      idToken,
				InitialAccessToken:  accessToken,
				InitialRefreshToken: refreshToken,
				CurrentIDToken:      idToken,
				CurrentAccessToken:  accessToken,
				CurrentRefreshToken: refreshToken,
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

func OIDCUniqueID(provider, subject string) string {
	return fmt.Sprintf("%s:%s", provider, subject)
}

func (c *CredentialsOIDC) GetProvider(provider, subject string) (k int, found bool) {
	for k, p := range c.Providers {
		if p.Subject == subject && p.Provider == provider {
			return k, true
		}
	}
	return -1, false
}
