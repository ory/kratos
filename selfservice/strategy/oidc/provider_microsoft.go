// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/x/httpx"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

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
	if len(strings.TrimSpace(m.config.Tenant)) == 0 {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("No Tenant specified for the `microsoft` oidc provider %s", m.config.ID))
	}

	endpointPrefix := "https://login.microsoftonline.com/" + m.config.Tenant
	endpoint := oauth2.Endpoint{
		AuthURL:  endpointPrefix + "/oauth2/v2.0/authorize",
		TokenURL: endpointPrefix + "/oauth2/v2.0/token",
	}

	return m.oauth2ConfigFromEndpoint(ctx, endpoint), nil
}

func (m *ProviderMicrosoft) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
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
	if m.config.SubjectSource != "me" {
		return claims, nil
	}

	o, err := m.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	if claims.RawClaims == nil {
		claims.RawClaims = map[string]interface{}{}
	}

	ctx, client := httpx.SetOAuth2(ctx, m.reg.HTTPClient(ctx), o, exchange)

	// Query to request most user fields from the Graph API (User.Read scope)
	// https://learn.microsoft.com/en-us/previous-versions/azure/ad/graph/api/entity-and-complex-type-reference#user-entity
	// Nota bene: some fields are returned from the API when no $select is specified which are not listed in the above link!
	// The query below was constructed by using the Graph Explorer and comparing against a query without $select.
	// https://developer.microsoft.com/en-us/graph/graph-explorer?request=me%3F%24select%3Did%2CbusinessPhones%2CaccountEnabled%2Ccity%2Ccountry%2CcreationType%2CdeletionTimestamp%2Cdepartment%2CdirSyncEnabled%2CdisplayName%2CemployeeId%2CfacsimileTelephoneNumber%2CgivenName%2CimmutableId%2CjobTitle%2ClastDirSyncTime%2Cmail%2CmobilePhone%2CofficeLocation%2CmailNickname%2Cmobile%2CobjectId%2CobjectType%2ConPremisesSecurityIdentifier%2CotherMails%2CpasswordPolicies%2CpasswordProfile%2CphysicalDeliveryOfficeName%2CpostalCode%2CpreferredLanguage%2CproxyAddresses%2CrefreshTokensValidFromDateTime%2CshowInAddressList%2CsignInNames%2CsipProxyAddress%2Cstate%2CstreetAddress%2Csurname%2CtelephoneNumber%2CthumbnailPhoto%2CusageLocation%2CuserIdentities%2CuserPrincipalName%2CuserType&method=GET&version=v1.0&GraphUrl=https://graph.microsoft.com
	query := "?$select=id,businessPhones,accountEnabled,city,country,creationType,deletionTimestamp,department,dirSyncEnabled,displayName,employeeId,facsimileTelephoneNumber,givenName,immutableId,jobTitle,lastDirSyncTime,mail,mobilePhone,officeLocation,mailNickname,mobile,objectId,objectType,onPremisesSecurityIdentifier,otherMails,passwordPolicies,passwordProfile,physicalDeliveryOfficeName,postalCode,preferredLanguage,proxyAddresses,refreshTokensValidFromDateTime,showInAddressList,signInNames,sipProxyAddress,state,streetAddress,surname,telephoneNumber,thumbnailPhoto,usageLocation,userIdentities,userPrincipalName,userType"
	claims.Subject, claims.RawClaims["user"], err = m.subjectFromGraphAPI(ctx, client, query)
	if err != nil {
		m.reg.Logger().WithError(err).Error("Unable to fetch user from Microsoft Graph API with $select. Attempting fallback without $select.")
		// fallback code path when the MS GraphAPI does not respond as expected
		query = ""
		claims.Subject, claims.RawClaims["user"], err = m.subjectFromGraphAPI(ctx, client, query)
		if err != nil {
			return nil, err
		}
	}

	return claims, nil
}

func (m *ProviderMicrosoft) subjectFromGraphAPI(ctx context.Context, client *retryablehttp.Client, query string) (subject string, rawClaims map[string]any, err error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", "https://graph.microsoft.com/v1.0/me"+query, nil)
	if err != nil {
		return "", nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to fetch from `https://graph.microsoft.com/v1.0/me`: %s", err))
	}
	defer resp.Body.Close()

	if err := logUpstreamError(m.reg.Logger(), resp); err != nil {
		return "", nil, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawClaims); err != nil {
		return "", nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to decode JSON from `https://graph.microsoft.com/v1.0/me`: %s", err))
	}

	trace.SpanFromContext(ctx).SetAttributes(attribute.StringSlice("microsoft.user.claims", maps.Keys(rawClaims)))

	ok := false
	if subject, ok = rawClaims["id"].(string); !ok {
		return "", nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to retrieve subject from `https://graph.microsoft.com/v1.0/me`: %s", err))
	}
	return subject, rawClaims, nil
}

type microsoftUnverifiedClaims struct {
	TenantID string `json:"tid,omitempty"`
}

func (c *microsoftUnverifiedClaims) Valid() error {
	return nil
}
