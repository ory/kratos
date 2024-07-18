// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"

	"github.com/ory/herodot"

	"github.com/ory/x/urlx"
)

type Configuration struct {
	// ID is the provider's ID
	ID string `json:"id"`

	// Provider is either "generic" for a generic OAuth 2.0 / OpenID Connect Provider or one of:
	// - generic
	// - google
	// - github
	// - github-app
	// - gitlab
	// - microsoft
	// - discord
	// - salesforce
	// - slack
	// - facebook
	// - auth0
	// - vk
	// - yandex
	// - apple
	// - spotify
	// - netid
	// - dingtalk
	// - linkedin
	// - patreon
	Provider string `json:"provider"`

	// Label represents an optional label which can be used in the UI generation.
	Label string `json:"label"`

	// ClientID is the application's Client ID.
	ClientID string `json:"client_id"`

	// ClientSecret is the application's secret.
	ClientSecret string `json:"client_secret"`

	// IssuerURL is the OpenID Connect Server URL. You can leave this empty if `provider` is not set to `generic`.
	// If set, neither `auth_url` nor `token_url` are required.
	IssuerURL string `json:"issuer_url"`

	// AuthURL is the authorize url, typically something like: https://example.org/oauth2/auth
	// Should only be used when the OAuth2 / OpenID Connect server is not supporting OpenID Connect Discovery and when
	// `provider` is set to `generic`.
	AuthURL string `json:"auth_url"`

	// TokenURL is the token url, typically something like: https://example.org/oauth2/token
	// Should only be used when the OAuth2 / OpenID Connect server is not supporting OpenID Connect Discovery and when
	// `provider` is set to `generic`.
	TokenURL string `json:"token_url"`

	// Tenant is the Azure AD Tenant to use for authentication, and must be set when `provider` is set to `microsoft`.
	// Can be either `common`, `organizations`, `consumers` for a multitenant application or a specific tenant like
	// `8eaef023-2b34-4da1-9baa-8bc8c9d6a490` or `contoso.onmicrosoft.com`.
	Tenant string `json:"microsoft_tenant"`

	// SubjectSource is a flag which controls from which endpoint the subject identifier is taken by microsoft provider.
	// Can be either `userinfo` or `me`.
	// If the value is `userinfo` then the subject identifier is taken from sub field of userinfo standard endpoint response.
	// If the value is `me` then the `id` field of https://graph.microsoft.com/v1.0/me response is taken as subject.
	// The default is `userinfo`.
	SubjectSource string `json:"subject_source"`

	// TeamId is the Apple Developer Team ID that's needed for the `apple` `provider` to work.
	// It can be found Apple Developer website and combined with `apple_private_key` and `apple_private_key_id`
	// is used to generate `client_secret`
	TeamId string `json:"apple_team_id"`

	// PrivateKeyId is the private Apple key identifier. Keys can be generated via developer.apple.com.
	// This key should be generated with the `Sign In with Apple` option checked.
	// This is needed when `provider` is set to `apple`
	PrivateKeyId string `json:"apple_private_key_id"`

	// PrivateKeyId is the Apple private key identifier that can be downloaded during key generation.
	// This is needed when `provider` is set to `apple`
	PrivateKey string `json:"apple_private_key"`

	// Scope specifies optional requested permissions.
	Scope []string `json:"scope"`

	// Mapper specifies the JSONNet code snippet which uses the OpenID Connect Provider's data (e.g. GitHub or Google
	// profile information) to hydrate the identity's data.
	//
	// It can be either a URL (file://, http(s)://, base64://) or an inline JSONNet code snippet.
	Mapper string `json:"mapper_url"`

	// RequestedClaims is a string encoded json object that specifies claims and optionally their properties that should be
	// included in the id_token or returned from the UserInfo Endpoint.
	//
	// More information: https://openid.net/specs/openid-connect-core-1_0.html#ClaimsParameter
	RequestedClaims json.RawMessage `json:"requested_claims"`

	// An optional organization ID that this provider belongs to.
	// This parameter is only effective in the Ory Network.
	OrganizationID string `json:"organization_id"`

	// AdditionalIDTokenAudiences is a list of additional audiences allowed in the ID Token.
	// This is only relevant in OIDC flows that submit an IDToken instead of using the callback from the OIDC provider.
	AdditionalIDTokenAudiences []string `json:"additional_id_token_audiences"`

	// ClaimsSource is a flag which controls where the claims are taken from when
	// using the generic provider. Can be either `userinfo` (calls the userinfo
	// endpoint to get the claims) or `id_token` (takes the claims from the id
	// token). It defaults to `id_token`.
	ClaimsSource string `json:"claims_source"`
}

func (p Configuration) Redir(public *url.URL) string {
	if p.OrganizationID != "" {
		route := RouteOrganizationCallback
		route = strings.Replace(route, ":provider", p.ID, 1)
		route = strings.Replace(route, ":organization", p.OrganizationID, 1)
		return urlx.AppendPaths(public, route).String()
	}

	return urlx.AppendPaths(public, strings.Replace(RouteCallback, ":provider", p.ID, 1)).String()
}

type ConfigurationCollection struct {
	BaseRedirectURI string          `json:"base_redirect_uri"`
	Providers       []Configuration `json:"providers"`
}

// !!! WARNING !!!
//
// If you add a provider here, please also add a test to
// provider_private_net_test.go
var supportedProviders = map[string]func(config *Configuration, reg Dependencies) Provider{
	"generic":     NewProviderGenericOIDC,
	"google":      NewProviderGoogle,
	"github":      NewProviderGitHub,
	"github-app":  NewProviderGitHubApp,
	"gitlab":      NewProviderGitLab,
	"microsoft":   NewProviderMicrosoft,
	"discord":     NewProviderDiscord,
	"salesforce":  NewProviderSalesforce,
	"slack":       NewProviderSlack,
	"facebook":    NewProviderFacebook,
	"auth0":       NewProviderAuth0,
	"vk":          NewProviderVK,
	"yandex":      NewProviderYandex,
	"apple":       NewProviderApple,
	"spotify":     NewProviderSpotify,
	"netid":       NewProviderNetID,
	"dingtalk":    NewProviderDingTalk,
	"linkedin":    NewProviderLinkedIn,
	"linkedin_v2": NewProviderLinkedInV2,
	"patreon":     NewProviderPatreon,
	"lark":        NewProviderLark,
	"x":           NewProviderX,
}

func (c ConfigurationCollection) Provider(id string, reg Dependencies) (Provider, error) {
	for k := range c.Providers {
		p := c.Providers[k]
		if p.ID == id {
			if f, ok := supportedProviders[p.Provider]; ok {
				return f(&p, reg), nil
			}

			return nil, errors.Errorf("provider type %s is not supported, supported are: %v", p.Provider, maps.Keys(supportedProviders))
		}
	}
	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf(`OpenID Connect Provider "%s" is unknown or has not been configured`, id))
}
