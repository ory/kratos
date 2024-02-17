// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oid2

import "context"

type ProviderGenericOid2 struct {
	config *Configuration
	reg    Dependencies
}

func NewProviderGenericOid2(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderGenericOid2{
		config: config,
		reg:    reg,
	}
}

func (g *ProviderGenericOid2) Config() *Configuration {
	return g.config
}

func (g *ProviderGenericOid2) GetRedirectUrl(ctx context.Context) string {
	return g.config.Redir(g.reg.Config().Oid2RedirectURIBase(ctx))
}
