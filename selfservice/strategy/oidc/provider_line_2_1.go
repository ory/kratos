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
	return &ProviderLineV21{
		&ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (g *ProviderLineV21) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	o, err := g.ProviderGenericOIDC.OAuth2(ctx)
	// Line login requires adding id_token_key_type=JWK when getting the token in order to issue an HS256 token.
	opts = append(opts, oauth2.SetAuthURLParam("id_token_key_type", "JWK"))

	token, err := o.Exchange(ctx, code, opts...)

	return token, err

}
