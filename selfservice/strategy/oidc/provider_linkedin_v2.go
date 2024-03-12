// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/url"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type ProviderLinkedInV2 struct {
	*ProviderGenericOIDC
}

func NewProviderLinkedInV2(
	config *Configuration,
	reg Dependencies,
) Provider {
	config.ClaimsSource = ClaimsSourceUserInfo
	config.IssuerURL = "https://www.linkedin.com/oauth"

	return &ProviderLinkedInV2{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (l *ProviderLinkedInV2) wrapCtx(ctx context.Context) context.Context {
	// We need to overwrite the issuer here because the discovery URL is under
	// `https://www.linkedin.com/oauth/.well-known/openid-configuration`, wherease
	// the issuer is `https://www.linkedin.com` (without the `/oauth`). This is
	// not conformant according to the OIDC spec, but needed for LinkedIn.
	return gooidc.InsecureIssuerURLContext(ctx, "https://www.linkedin.com")
}

func (l *ProviderLinkedInV2) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return l.ProviderGenericOIDC.OAuth2(l.wrapCtx(ctx))
}

func (l *ProviderLinkedInV2) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	return l.ProviderGenericOIDC.Claims(l.wrapCtx(ctx), exchange, query)
}
