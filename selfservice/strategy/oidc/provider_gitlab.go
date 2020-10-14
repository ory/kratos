package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

const (
	defaultEndpoint = "https://gitlab.com"
)

type ProviderGitLab struct {
	*ProviderGenericOIDC
}

func NewProviderGitLab(
	config *Configuration,
	public *url.URL,
) *ProviderGitLab {
	return &ProviderGitLab{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			public: public,
		},
	}
}

func (g *ProviderGitLab) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := o.Client(ctx, exchange)

	u, err := g.endpoint()
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	u.Path = path.Join(u.Path, "/oauth/userinfo")
	req, err := http.NewRequest("GET", u.String(), nil)
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

func (g *ProviderGitLab) endpoint() (*url.URL, error) {
	var e = defaultEndpoint
	if len(g.config.IssuerURL) > 0 {
		e = g.config.IssuerURL
	}
	return url.Parse(e)
}
