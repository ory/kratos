package oidc

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ory/herodot"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"

	"github.com/slack-go/slack"
)

type ProviderSlack struct {
	config *Configuration
	public *url.URL
}

func NewProviderSlack(
	config *Configuration,
	public *url.URL,
) *ProviderSlack {
	return &ProviderSlack{
		config: config,
		public: public,
	}
}

func (d *ProviderSlack) Config() *Configuration {
	return d.config
}

func (d *ProviderSlack) oauth2() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     d.config.ClientID,
		ClientSecret: d.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			// slack's oauth v2 does not implement the oauth2 standard so we use the old version.
			// to use v2 we would need to rename the request 'scope' field to 'user_scope'.
			AuthURL:  "https://slack.com/oauth/authorize",
			TokenURL: slack.APIURL + "oauth.access",
		},
		RedirectURL: d.config.Redir(d.public),
		Scopes:      d.config.Scope,
	}
}

func (d *ProviderSlack) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return d.oauth2(), nil
}

func (d *ProviderSlack) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (d *ProviderSlack) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	grantedScopes := stringsx.Splitx(fmt.Sprintf("%s", exchange.Extra("scope")), ",")
	for _, check := range d.Config().Scope {
		if !stringslice.Has(grantedScopes, check) {
			return nil, errors.WithStack(ErrScopeMissing)
		}
	}

	api := slack.New(exchange.AccessToken)
	identity, err := api.GetUserIdentity()
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	claims := &Claims{
		Issuer:            "https://slack.com/oauth/",
		Subject:           identity.User.ID,
		Name:              identity.User.Name,
		PreferredUsername: identity.User.Name,
		Nickname:          identity.User.Name,
		Email:             identity.User.Email,
		EmailVerified:     true,
		Picture:           identity.User.Image512,
		Team:              identity.Team.ID,
	}

	return claims, nil
}
