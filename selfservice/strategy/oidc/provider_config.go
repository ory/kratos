package oidc

import (
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
	Provider string `json:"provider"`

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
	RequestedClaims string `json:"requested_claims"`
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
			switch p.Provider {
			case "generic":
				return NewProviderGenericOIDC(&p, public), nil
			case "google":
				return NewProviderGoogle(&p, public), nil
			case "github":
				return NewProviderGitHub(&p, public), nil
			case "microsoft":
				return NewProviderMicrosoft(&p, public), nil
			}
			return nil, errors.Errorf("provider type %s is not supported, supported are: %v", p.Provider, []string{"generic", "google", "github", "microsoft"})
		}
	}
	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf(`OpenID Connect Provider "%s" is unknown or has not been configured`, id))
}
