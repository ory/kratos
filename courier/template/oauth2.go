// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package template

import (
	hydraclientgo "github.com/ory/hydra-client-go/v2"
)

type (
	// OAuth2LoginRequest is a scoped view of the OAuth2 login request that
	// initiated a self-service flow. It is embedded into courier template data
	// so that messages can be branded per OAuth2 client without exposing the
	// full flow object.
	OAuth2LoginRequest struct {
		Challenge string       `json:"challenge"`
		Client    OAuth2Client `json:"client"`
	}

	// OAuth2Client contains the branding-relevant, admin-controlled fields of
	// the OAuth2 client that initiated a self-service flow.
	OAuth2Client struct {
		ClientID   string `json:"client_id"`
		ClientName string `json:"client_name,omitempty"`
		ClientURI  string `json:"client_uri,omitempty"`
		LogoURI    string `json:"logo_uri,omitempty"`
		Metadata   any    `json:"metadata,omitempty"`
	}
)

// NewOAuth2LoginRequest maps the Hydra login request to the scoped view embedded
// into courier template data. Only branding-relevant, admin-controlled client
// fields are exposed to messaging channels.
func NewOAuth2LoginRequest(hlr *hydraclientgo.OAuth2LoginRequest) *OAuth2LoginRequest {
	if hlr == nil {
		return nil
	}
	return &OAuth2LoginRequest{
		Challenge: hlr.Challenge,
		Client: OAuth2Client{
			ClientID:   hlr.Client.GetClientId(),
			ClientName: hlr.Client.GetClientName(),
			ClientURI:  hlr.Client.GetClientUri(),
			LogoURI:    hlr.Client.GetLogoUri(),
			Metadata:   hlr.Client.GetMetadata(),
		},
	}
}
