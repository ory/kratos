// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/url"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"
)

type ProviderJackson struct {
	*ProviderGenericOIDC
}

func NewProviderJackson(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderJackson{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (j *ProviderJackson) setProvider(ctx context.Context) {
	if j.ProviderGenericOIDC.p == nil {
		config := gooidc.ProviderConfig{
			IssuerURL:     j.config.IssuerURL,
			AuthURL:       j.config.AuthURL,
			TokenURL:      j.config.TokenURL,
			DeviceAuthURL: "",
			UserInfoURL:   j.config.IssuerURL + "/api/oauth/userinfo",
			JWKSURL:       j.config.IssuerURL + "/oauth/jwks",
			Algorithms:    []string{"RS256"},
		}
		j.ProviderGenericOIDC.p = config.NewProvider(j.withHTTPClientContext(ctx))
	}
}

func (j *ProviderJackson) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	j.setProvider(ctx)
	endpoint := j.ProviderGenericOIDC.p.Endpoint()
	config := j.oauth2ConfigFromEndpoint(ctx, endpoint)
	config.RedirectURL = urlx.AppendPaths(
		j.reg.Config().SAMLRedirectURIBase(ctx),
		"/self-service/methods/saml/callback/"+j.config.ID).String()

	return j.ProviderGenericOIDC.OAuth2(ctx)
}

func (j *ProviderJackson) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	j.setProvider(ctx)
	return j.claimsFromIDToken(ctx, exchange)
}

func (j *ProviderJackson) claimsFromIDToken(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	p, raw, err := j.idTokenAndProvider(ctx, exchange)
	if err != nil {
		return nil, err
	}

	return j.verifyAndDecodeClaimsWithProvider(ctx, p, raw)
}

func (j *ProviderJackson) verifyAndDecodeClaimsWithProvider(ctx context.Context, provider *gooidc.Provider, raw string) (*Claims, error) {
	verifier := provider.VerifierContext(j.withHTTPClientContext(ctx), &gooidc.Config{
		ClientID:        j.config.ClientID,
		SkipIssuerCheck: true,
	})
	token, err := verifier.Verify(ctx, raw)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
	}

	var claims Claims
	if err := token.Claims(&claims); err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
	}

	var rawClaims map[string]interface{}
	if err := token.Claims(&rawClaims); err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("%s", err))
	}
	claims.RawClaims = rawClaims

	return &claims, nil
}
