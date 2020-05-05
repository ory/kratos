package oidc

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/x/urlx"
)

type Configuration struct {
	// RequestID is the provider RequestID
	ID string `json:"id"`

	// Provider is either "generic" for a generic OAuth 2.0 / OpenID Connect Provider or one of:
	// - generic
	// - google
	Provider string `json:"provider"`

	// ClientID is the application's RequestID.
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

	// Scope specifies optional requested permissions.
	Scope []string `json:"scope"`

	// Normalize specifies the JSONNet code snippet which uses the OpenID Connect Provider's data (e.g. GitHub or Google
	// profile information) to hydrate the identity's data.
	Normalize string `json:"normalizer"`
}

func (p Configuration) Redir(public *url.URL) string {
	return urlx.AppendPaths(public,
		strings.Replace(CallbackPath, ":provider", p.ID, 1),
	).String()
}

type ConfigurationCollection struct {
	Providers []Configuration `json:"providers"`
}

func (c ConfigurationCollection) Provider(id string, public *url.URL) (Provider, error) {
	for _, p := range c.Providers {
		if p.ID == id {
			switch p.Provider {
			case "generic":
				return NewProviderGenericOIDC(&p, public), nil
			case "google":
				return NewProviderGoogle(&p, public), nil
			case "github":
				return NewProviderGitHub(&p, public), nil
			}
			return nil, errors.Errorf("provider type %s is not supported, supported are: %v", p.Provider, []string{"google"})
		}
	}
	return nil, errors.WithStack(herodot.ErrNotFound.WithReasonf(`OpenID Connect Provider "%s" is unknown or has not been configured`, id))
}
