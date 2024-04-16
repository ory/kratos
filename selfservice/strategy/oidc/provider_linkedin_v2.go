// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

func NewProviderLinkedInV2(
	config *Configuration,
	reg Dependencies,
) Provider {
	config.ClaimsSource = ClaimsSourceUserInfo
	config.IssuerURL = "https://www.linkedin.com/oauth"

	return &ProviderGenericOIDC{
		config: config,
		reg:    reg,
	}
}
