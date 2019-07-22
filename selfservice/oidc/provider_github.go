package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"

	"github.com/ory/herodot"
)

type ProviderGitHub struct {
	config *Configuration
	public *url.URL
}

func NewProviderGitHub(
	config *Configuration,
	public *url.URL,
) *ProviderGitHub {
	return &ProviderGitHub{
		config: config,
		public: public,
	}
}

func (g *ProviderGitHub) Config() *Configuration {
	return g.config
}

func (g *ProviderGitHub) oauth2() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint:     github.Endpoint,
		Scopes:       g.config.Scope,
		RedirectURL:  g.config.Redir(g.public),
	}
}

func (g *ProviderGitHub) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(), nil
}

func (g *ProviderGitHub) nr(client *http.Client, path string, out interface{}) error {
	req, err := http.NewRequest("GET", "https://api.github.com"+path, nil)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	res, err := client.Do(req)
	if err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected status code %d from userinfo endpoint but got %d", http.StatusOK, res.StatusCode))
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	return nil
}

func (g *ProviderGitHub) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	grantedScopes := stringsx.Splitx(fmt.Sprintf("%s", exchange.Extra("scope")), ",")
	for _, check := range g.Config().Scope {
		if !stringslice.Has(grantedScopes, check) {
			return nil, errors.WithStack(ErrScopeMissing)
		}
	}

	client := g.oauth2().Client(ctx, exchange)

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	var claims json.RawMessage
	if err := g.nr(client, "/user", &claims); err != nil {
		return nil, err
	}

	subject := fmt.Sprintf("%d", gjson.GetBytes(claims, "id").Int())
	if len(subject) == 0 {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected the user id to be defined but it is empty"))
	}

	// GitHub does not provide the user's private emails in the call to `/user`. Therefore, if scope "user:email" is set,
	// we want to make another request to `/user/emails` and merge that with our claims.
	if stringslice.Has(grantedScopes, "user:email") {
		var emails json.RawMessage
		if err := g.nr(client, "/user/emails", &emails); err != nil {
			return nil, err
		}

		claims, err = sjson.SetRawBytes(claims, "emails", emails)
		if len(subject) == 0 {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to set emails claims: %s", err))
		}
	}

	return &Claims{
		Subject: subject,
		Traits:  claims,
	}, nil
}
