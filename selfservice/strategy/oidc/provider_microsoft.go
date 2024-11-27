// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
)

var _ OAuth2Provider = (*ProviderMicrosoft)(nil)

type ProviderMicrosoft struct {
	*ProviderGenericOIDC
}

func NewProviderMicrosoft(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderMicrosoft{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (m *ProviderMicrosoft) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	if strings.TrimSpace(m.config.Tenant) == "" {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("No Tenant specified for the `microsoft` oidc provider %s", m.config.ID))
	}

	endpointPrefix := "https://login.microsoftonline.com/" + m.config.Tenant
	endpoint := oauth2.Endpoint{
		AuthURL:  endpointPrefix + "/oauth2/v2.0/authorize",
		TokenURL: endpointPrefix + "/oauth2/v2.0/token",
	}

	return m.oauth2ConfigFromEndpoint(ctx, endpoint), nil
}

func (m *ProviderMicrosoft) Claims(ctx context.Context, exchange *oauth2.Token, _ url.Values) (*Claims, error) {
	raw, ok := exchange.Extra("id_token").(string)
	if !ok || len(raw) == 0 {
		return nil, errors.WithStack(ErrIDTokenMissing)
	}

	parser := new(jwt.Parser)
	unverifiedClaims := microsoftUnverifiedClaims{}
	if _, _, err := parser.ParseUnverified(raw, &unverifiedClaims); err != nil {
		return nil, err
	}

	if _, err := uuid.FromString(unverifiedClaims.TenantID); err != nil {
		return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("TenantID claim is not a valid UUID: %s", err))
	}

	issuer := "https://login.microsoftonline.com/" + unverifiedClaims.TenantID + "/v2.0"
	ctx = context.WithValue(ctx, oauth2.HTTPClient, m.reg.HTTPClient(ctx).HTTPClient)
	p, err := gooidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to initialize OpenID Connect Provider: %s", err))
	}

	claims, err := m.verifyAndDecodeClaimsWithProvider(ctx, p, raw)
	if err != nil {
		return nil, err
	}

	return m.updateSubject(ctx, claims, exchange)
}

func (m *ProviderMicrosoft) updateSubject(ctx context.Context, claims *Claims, exchange *oauth2.Token) (*Claims, error) {
	if m.config.SubjectSource == "me" {
		o, err := m.OAuth2(ctx)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
		}

		ctx, client := httpx.SetOAuth2(ctx, m.reg.HTTPClient(ctx), o, exchange)
		req, err := retryablehttp.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/me", nil)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to fetch from `https://graph.microsoft.com/v1.0/me`: %s", err))
		}
		defer resp.Body.Close()

		if err := logUpstreamError(m.reg.Logger(), resp); err != nil {
			return nil, err
		}

		var user struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to decode JSON from `https://graph.microsoft.com/v1.0/me`: %s", err))
		}

		claims.Subject = user.ID
	}

	if m.config.SubjectSource == "oid" {
		claims.Subject = claims.Object
	}

	return claims, nil
}

type microsoftUnverifiedClaims struct {
	TenantID string `json:"tid,omitempty"`
}

func (c *microsoftUnverifiedClaims) Valid() error {
	return nil
}
