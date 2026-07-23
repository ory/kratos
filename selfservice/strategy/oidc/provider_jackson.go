// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"
)

// jacksonTrustedIssuersEnv is a comma-separated list of token issuers accepted
// in addition to the configured issuer URL. It keeps ID tokens verifiable while
// the issuer URL migrates from one domain to another: set it to the old and new
// issuers for the duration of the migration, then unset it.
const jacksonTrustedIssuersEnv = "SAML_SP_TRUSTED_ISSUERS"

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
	if j.p == nil {
		internalHost := strings.TrimSuffix(j.config.TokenURL, "/api/oauth/token")
		config := oidc.ProviderConfig{
			IssuerURL:     j.config.IssuerURL,
			AuthURL:       j.config.AuthURL,
			TokenURL:      j.config.TokenURL,
			DeviceAuthURL: "",
			UserInfoURL:   internalHost + "/api/oauth/userinfo",
			JWKSURL:       internalHost + "/oauth/jwks",
			Algorithms:    []string{"RS256"},
		}
		j.p = config.NewProvider(j.withHTTPClientContext(ctx))
	}
}

func (j *ProviderJackson) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	j.setProvider(ctx)
	endpoint := j.p.Endpoint()
	config := j.oauth2ConfigFromEndpoint(ctx, endpoint)
	config.RedirectURL = urlx.AppendPaths(
		j.reg.Config().SAMLRedirectURIBase(ctx),
		"/self-service/methods/saml/callback/"+j.config.ID,
	).String()

	return config, nil
}

func (j *ProviderJackson) Claims(ctx context.Context, exchange *oauth2.Token, _ url.Values) (*Claims, error) {
	j.setProvider(ctx)

	raw, ok := exchange.Extra("id_token").(string)
	if !ok || len(raw) == 0 {
		return nil, errors.WithStack(ErrIDTokenMissing())
	}

	// go-oidc's SkipIssuerCheck defers issuer validation to the caller. The
	// issuer is verified below against the configured issuer plus the
	// JacksonTrustedIssuersEnv allowlist.
	token, err := j.p.VerifierContext(
		j.withHTTPClientContext(ctx),
		&oidc.Config{ClientID: j.config.ClientID, SkipIssuerCheck: true},
	).Verify(ctx, raw)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest().WithReasonf("%s", err))
	}

	if token.Issuer != j.config.IssuerURL && !slices.Contains(jacksonTrustedIssuers(), token.Issuer) {
		return nil, errors.WithStack(herodot.ErrBadRequest().WithReasonf(
			"oidc: id token issued by a different provider, expected %q got %q", j.config.IssuerURL, token.Issuer,
		))
	}

	var claims Claims
	if err := token.Claims(&claims); err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest().WithReasonf("%s", err))
	}

	var rawClaims map[string]interface{}
	if err := token.Claims(&rawClaims); err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest().WithReasonf("%s", err))
	}
	claims.RawClaims = rawClaims

	return &claims, nil
}

func jacksonTrustedIssuers() []string {
	var issuers []string
	for iss := range strings.SplitSeq(os.Getenv(jacksonTrustedIssuersEnv), ",") {
		if iss = strings.TrimSpace(iss); iss != "" {
			issuers = append(issuers, iss)
		}
	}
	return issuers
}
