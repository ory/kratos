package oidc

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/x/urlx"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
)

const (
	defaultBrokerScheme = "https"
	defaultBrokerHost   = "broker.netid.de"
)

type ProviderNetID struct {
	*ProviderGenericOIDC
}

func NewProviderNetID(
	config *Configuration,
	reg dependencies,
) *ProviderNetID {
	return &ProviderNetID{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (n *ProviderNetID) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return n.oAuth2(ctx)
}

func (n *ProviderNetID) oAuth2(ctx context.Context) (*oauth2.Config, error) {
	u := n.brokerURL()

	authURL := urlx.AppendPaths(u, "/authorize")
	tokenURL := urlx.AppendPaths(u, "/token")

	return &oauth2.Config{
		ClientID:     n.config.ClientID,
		ClientSecret: n.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL.String(),
			TokenURL: tokenURL.String(),
		},
		Scopes:      n.config.Scope,
		RedirectURL: n.config.Redir(n.reg.Config(ctx).OIDCRedirectURIBase()),
	}, nil

}

func (n *ProviderNetID) Claims(ctx context.Context, exchange *oauth2.Token, _ url.Values) (*Claims, error) {
	o, err := n.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := n.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs(), httpx.ResilientClientWithClient(o.Client(ctx, exchange)))

	u := n.brokerURL()
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	userInfoURL := urlx.AppendPaths(u, "/userinfo")
	req, err := retryablehttp.NewRequest("GET", userInfoURL.String(), nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	var claims Claims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	return &claims, nil
}

func (n *ProviderNetID) brokerURL() *url.URL {
	return &url.URL{Scheme: defaultBrokerScheme, Host: defaultBrokerHost}
}
