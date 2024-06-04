// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/workos/workos-go/v3/pkg/sso"

	_ "github.com/motemen/go-loghttp/global"

	"github.com/ory/herodot"
)

type ProviderWorkOS struct {
	config    *Configuration
	reg       Dependencies
	ssoClient *sso.Client
}

func NewProviderWorkOS(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderWorkOS{
		config: config,
		reg:    reg,
		ssoClient: &sso.Client{
			APIKey:   config.ClientSecret,
			ClientID: config.ClientID,
		},
	}
}

func (g *ProviderWorkOS) Config() *Configuration {
	return g.config
}

func (g *ProviderWorkOS) oauth2(ctx context.Context) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://api.workos.com/sso/authorize?organization=" + g.config.WorkOSOrganizationId,
		TokenURL: "https://api.workos.com/sso/token",
	}

	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint:     endpoint,
		RedirectURL:  g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (g *ProviderWorkOS) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx), nil
}

func (g *ProviderWorkOS) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (g *ProviderWorkOS) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	profile, err := g.ssoClient.GetProfile(
		ctx,
		sso.GetProfileOpts{
			AccessToken: exchange.AccessToken,
		},
	)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	claims := &Claims{
		Subject:    profile.ID,
		GivenName:  profile.FirstName,
		FamilyName: profile.LastName,
		Email:      profile.Email,
		Issuer:     "https://api.workos.com/sso/authorize?organization=" + g.config.WorkOSOrganizationId,
	}

	return claims, nil
}
