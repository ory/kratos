package oidc

import (
	"context"
	"encoding/json"
	"net/url"
	"path"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
)

const (
	defaultBrokerEndpoint = "https://broker.netid.de"
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
	endpoint, err := n.endpoint()
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	authUrl := *endpoint
	tokenUrl := *endpoint

	authUrl.Path = path.Join(authUrl.Path, "/authorize")
	tokenUrl.Path = path.Join(tokenUrl.Path, "/token")

	return &oauth2.Config{
		ClientID:     n.config.ClientID,
		ClientSecret: n.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authUrl.String(),
			TokenURL: tokenUrl.String(),
		},
		Scopes:      n.config.Scope,
		RedirectURL: n.config.Redir(n.reg.Config(ctx).OIDCRedirectURIBase()),
	}, nil

}

func (n *ProviderNetID) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	o, err := n.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := n.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs(), httpx.ResilientClientWithClient(o.Client(ctx, exchange)))

	u, err := n.endpoint()
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	u.Path = path.Join(u.Path, "/userinfo")
	req, err := retryablehttp.NewRequest("GET", u.String(), nil)
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

func (n *ProviderNetID) endpoint() (*url.URL, error) {
	return url.Parse(defaultBrokerEndpoint)
}
