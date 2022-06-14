package oidc

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ory/kratos/x"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"

	ghapi "github.com/google/go-github/v38/github"

	"github.com/ory/herodot"
)

type ProviderGitHub struct {
	config *Configuration
	reg    dependencies
}

func NewProviderGitHub(
	config *Configuration,
	reg dependencies,
) *ProviderGitHub {
	return &ProviderGitHub{
		config: config,
		reg:    reg,
	}
}

func (g *ProviderGitHub) Config() *Configuration {
	return g.config
}

func (g *ProviderGitHub) oauth2(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint:     github.Endpoint,
		Scopes:       g.config.Scope,
		RedirectURL:  g.config.Redir(g.reg.Config(ctx).OIDCRedirectURIBase()),
	}
}

func (g *ProviderGitHub) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx), nil
}

func (g *ProviderGitHub) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (g *ProviderGitHub) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	grantedScopes := stringsx.Splitx(fmt.Sprintf("%s", exchange.Extra("scope")), ",")
	for _, check := range g.Config().Scope {
		if !stringslice.Has(grantedScopes, check) {
			return nil, errors.WithStack(ErrScopeMissing)
		}
	}

	gh := ghapi.NewClient(g.oauth2(ctx).Client(ctx, exchange))

	user, _, err := gh.Users.Get(ctx, "")
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	claims := &Claims{
		Subject:   fmt.Sprintf("%d", user.GetID()),
		Issuer:    github.Endpoint.TokenURL,
		Name:      user.GetName(),
		Nickname:  user.GetLogin(),
		Website:   user.GetBlog(),
		Picture:   user.GetAvatarURL(),
		Profile:   user.GetHTMLURL(),
		UpdatedAt: user.GetUpdatedAt().Unix(),
	}

	// GitHub does not provide the user's private emails in the call to `/user`. Therefore, if scope "user:email" is set,
	// we want to make another request to `/user/emails` and merge that with our claims.
	if stringslice.Has(grantedScopes, "user:email") {
		emails, _, err := gh.Users.ListEmails(ctx, nil)
		if err != nil {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
		}

		for k, e := range emails {
			// If it is the primary email or it's the last email (no primary email set?), set the email.
			if e.GetPrimary() || k == len(emails) {
				claims.Email = e.GetEmail()
				claims.EmailVerified = x.ConvertibleBoolean(e.GetVerified())
				break
			}
		}
	}

	return claims, nil
}
