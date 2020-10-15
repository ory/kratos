package oidc

import (
	"context"
	"fmt"
	"net/url"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"
)

type ProviderDiscord struct {
	config *Configuration
	public *url.URL
}

func NewProviderDiscord(
	config *Configuration,
	public *url.URL,
) *ProviderDiscord {
	return &ProviderDiscord{
		config: config,
		public: public,
	}
}

func (d *ProviderDiscord) Config() *Configuration {
	return d.config
}

func (d *ProviderDiscord) oauth2() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     d.config.ClientID,
		ClientSecret: d.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  discordgo.EndpointOauth2 + "authorize",
			TokenURL: discordgo.EndpointOauth2 + "token",
		},
		RedirectURL: d.config.Redir(d.public),
		Scopes:      d.config.Scope,
	}
}

func (d *ProviderDiscord) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return d.oauth2(), nil
}

func (d *ProviderDiscord) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	if isForced(r) {
		return []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("prompt", "consent"),
		}
	}
	return []oauth2.AuthCodeOption{}
}

func (d *ProviderDiscord) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	grantedScopes := stringsx.Splitx(fmt.Sprintf("%s", exchange.Extra("scope")), " ")
	for _, check := range d.Config().Scope {
		if !stringslice.Has(grantedScopes, check) {
			return nil, errors.WithStack(ErrScopeMissing)
		}
	}

	dg, err := discordgo.New(fmt.Sprintf("Bearer %s", exchange.AccessToken))
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	// TODO: upgrade github.com/bwmarrin/discordgo once it supports api v8: https://github.com/bwmarrin/discordgo/issues/822
	user, err := dg.User("@me")
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	claims := &Claims{
		Issuer:            discordgo.EndpointOauth2,
		Subject:           user.ID,
		Name:              fmt.Sprintf("%s#%s", user.Username, user.Discriminator),
		Nickname:          user.Username,
		PreferredUsername: user.Username,
		Picture:           user.AvatarURL(""),
		Email:             user.Email,
		EmailVerified:     user.Verified,
		Locale:            user.Locale,
	}

	return claims, nil
}
