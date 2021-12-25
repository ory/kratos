package oidc

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/pkg/errors"

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
	// - slack
	// - facebook
	// - vk
	// - yandex
	// - apple
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
	Tenant string `json:"tenant"`

	// TeamId is the Apple Developer Team ID that's needed for the `apple` `provider` to work.
	// It can be found Apple Developer website and combined with `private_key` and `private_key_id`
	// is used to generate `client_secret`
	TeamId string `json:"team_id"`

	// PrivateKeyId is the private Apple key identifier. Keys can be generated via developer.apple.com.
	// This key should be generated with the `Sign In with Apple` option checked.
	// This is needed when `provider` is set to `apple`
	PrivateKeyId string `json:"private_key_id"`

	// PrivateKeyId is the Apple private key identifier that can be downloaded during key generation.
	// This is needed when `provider` is set to `apple`
	PrivateKey string `json:"private_key"`

	// Scope specifies optional requested permissions.
	Scope []string `json:"scope"`

	// Mapper specifies the JSONNet code snippet which uses the OpenID Connect Provider's data (e.g. GitHub or Google
	// profile information) to hydrate the identity's data.
	//
	// It can be either a URL (file://, http(s)://, base64://) or an inline JSONNet code snippet.
	Mapper string `json:"mapper_url"`

	// RequestedClaims string encoded json object that specifies claims and optionally their properties which should be
	// included in the id_token or returned from the UserInfo Endpoint.
	//
	// More information: https://openid.net/specs/openid-connect-core-1_0.html#ClaimsParameter
	RequestedClaims json.RawMessage `json:"requested_claims"`
}

func (p Configuration) Redir(public *url.URL) string {
	return urlx.AppendPaths(public,
		strings.Replace(RouteCallback, ":provider", p.ID, 1),
	).String()
}

type ConfigurationCollection struct {
	Providers []Configuration `json:"providers"`
}

func (c ConfigurationCollection) Provider(id string, public *url.URL) (Provider, error) {
	for k := range c.Providers {
		p := c.Providers[k]
		if p.ID == id {
			var providerNames []string
			var addProviderName = func(pn string) string {
				providerNames = append(providerNames, pn)
				return pn
			}

			switch p.Provider {
			case addProviderName("generic"):
				return NewProviderGenericOIDC(&p, public), nil
			case addProviderName("google"):
				return NewProviderGoogle(&p, public), nil
			case addProviderName("github"):
				return NewProviderGitHub(&p, public), nil
			case addProviderName("github-app"):
				return NewProviderGitHubApp(&p, public), nil
			case addProviderName("gitlab"):
				return NewProviderGitLab(&p, public), nil
			case addProviderName("microsoft"):
				return NewProviderMicrosoft(&p, public), nil
			case addProviderName("discord"):
				return NewProviderDiscord(&p, public), nil
			case addProviderName("slack"):
				return NewProviderSlack(&p, public), nil
			case addProviderName("facebook"):
				return NewProviderFacebook(&p, public), nil
			case addProviderName("auth0"):
				return NewProviderAuth0(&p, public), nil
			case addProviderName("vk"):
				return NewProviderVK(&p, public), nil
			case addProviderName("yandex"):
				return NewProviderYandex(&p, public), nil
			case addProviderName("apple"):
				return NewProviderApple(&p, public), nil
			case addProviderName("spotify"):
				return NewProviderSpotify(&p, public), nil
			}
			return nil, errors.Errorf("provider type %s is not supported, supported are: %v", p.Provider, providerNames)
		}
	}
	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf(`OpenID Connect Provider "%s" is unknown or has not been configured`, id))
}
