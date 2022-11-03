// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

type ProviderGoogle struct {
	*ProviderGenericOIDC
}

func NewProviderGoogle(
	config *Configuration,
	reg dependencies,
) *ProviderGoogle {
	config.IssuerURL = "https://accounts.google.com"
	return &ProviderGoogle{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}
