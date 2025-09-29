// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"

	"golang.org/x/oauth2"
)

type ProviderLineV21 struct {
	*ProviderGenericOIDC
}

func NewProviderLineV21(
	config *Configuration,
	reg Dependencies,
) Provider {
	config.IssuerURL = "https://access.line.me"
	return &ProviderLineV21{
		&ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (g *ProviderLineV21) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, err
	}

	// Line login requires adding id_token_key_type=JWK when getting the token in order to issue an ES256 token.
	// https://blog.miniasp.com/post/2022/04/08/LINE-Login-with-OpenID-Connect-in-ASPNET-Core (Chinese)
	opts = append(opts, oauth2.SetAuthURLParam("id_token_key_type", "JWK"))
	return o.Exchange(ctx, code, opts...)
}
