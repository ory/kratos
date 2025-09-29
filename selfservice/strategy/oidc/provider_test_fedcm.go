// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

// ProviderTestFedcm is a mock provider to test FedCM.
type ProviderTestFedcm struct {
	*ProviderGenericOIDC
}

var _ OAuth2Provider = (*ProviderTestFedcm)(nil)

func NewProviderTestFedcm(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderTestFedcm{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (g *ProviderTestFedcm) Verify(_ context.Context, rawIDToken string) (claims *Claims, err error) {
	rawClaims := &struct {
		Claims
		jwt.MapClaims
	}{}
	_, err = jwt.ParseWithClaims(rawIDToken, rawClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(`xxxxxxx`), nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, err
	}
	rawClaims.Issuer = "https://example.com/fedcm"

	if err = rawClaims.Validate(); err != nil {
		return nil, err
	}

	return &rawClaims.Claims, nil
}
