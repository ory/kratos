// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oid2

type ProviderSteam struct {
	*ProviderGenericOid2
}

const SteamDiscoveryUrl = "https://steamcommunity.com/openid"

func NewProviderSteam(
	config *Configuration,
	reg Dependencies,
) Provider {
	config.DiscoveryUrl = SteamDiscoveryUrl

	return &ProviderSteam{
		ProviderGenericOid2: &ProviderGenericOid2{
			config: config,
			reg:    reg,
		},
	}
}
