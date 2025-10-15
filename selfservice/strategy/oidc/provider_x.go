// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ory/x/otelx"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
)

var _ OAuth1Provider = (*ProviderX)(nil)

const (
	xUserInfoBase      = "https://api.twitter.com/1.1/account/verify_credentials.json"
	xUserInfoWithEmail = xUserInfoBase + "?include_email=true"
)

type ProviderX struct {
	config *Configuration
	reg    Dependencies
}

func (p *ProviderX) Config() *Configuration {
	return p.config
}

func NewProviderX(
	config *Configuration,
	reg Dependencies,
) Provider {
	return &ProviderX{
		config: config,
		reg:    reg,
	}
}

func (p *ProviderX) ExchangeToken(ctx context.Context, req *http.Request) (*oauth1.Token, error) {
	requestToken, verifier, err := oauth1.ParseAuthorizationCallback(req)
	if err != nil {
		return nil, err
	}

	accessToken, accessSecret, err := p.OAuth1(ctx).AccessToken(requestToken, "", verifier)
	if err != nil {
		return nil, err
	}

	return oauth1.NewToken(accessToken, accessSecret), nil
}

func (p *ProviderX) AuthURL(ctx context.Context, state string) (_ string, err error) {
	ctx, span := p.reg.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.oidc.ProviderLinkedIn.fetch")
	defer otelx.End(span, &err)

	c := p.OAuth1(ctx)

	// We need to cheat so that callback validates on return
	c.CallbackURL = c.CallbackURL + fmt.Sprintf("?state=%s&code=unused", state)

	requestToken, _, err := c.RequestToken()
	if err != nil {
		span.RecordError(err)
		return "", errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf(`Unable to sign in with X because the OAuth1 request token could not be initialized: %s`, err))
	}

	authzURL, err := c.AuthorizationURL(requestToken)
	if err != nil {
		span.RecordError(err)
		return "", errors.WithStack(herodot.ErrMisconfiguration.WithWrap(err).WithReasonf(`Unable to sign in with X because the OAuth1 authorization URL could not be parsed: %s`, err))
	}

	return authzURL.String(), nil
}

func (p *ProviderX) CheckError(ctx context.Context, r *http.Request) error {
	if r.URL.Query().Get("denied") == "" {
		return nil
	}

	return errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to sign in with X because the user denied the request.`))
}

func (p *ProviderX) OAuth1(ctx context.Context) *oauth1.Config {
	return &oauth1.Config{
		ConsumerKey:    p.config.ClientID,
		ConsumerSecret: p.config.ClientSecret,
		Endpoint:       twitter.AuthenticateEndpoint,
		CallbackURL:    p.config.Redir(p.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (p *ProviderX) userInfoEndpoint() string {
	for _, scope := range p.config.Scope {
		if scope == "email" {
			return xUserInfoWithEmail
		}
	}

	return xUserInfoBase
}

func (p *ProviderX) Claims(ctx context.Context, token *oauth1.Token) (*Claims, error) {
	ctx = context.WithValue(ctx, oauth1.HTTPClient, p.reg.HTTPClient(ctx).HTTPClient)

	c := p.OAuth1(ctx)
	client := c.Client(ctx, token)
	endpoint := p.userInfoEndpoint()

	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}
	defer func() { _ = resp.Body.Close() }()

	if err := logUpstreamError(p.reg.Logger(), resp); err != nil {
		return nil, err
	}

	user := &xUser{}
	if err := json.NewDecoder(resp.Body).Decode(user); err != nil {
		return nil, errors.WithStack(herodot.ErrUpstreamError.WithWrap(err).WithReasonf("%s", err))
	}

	website := ""
	if user.URL != nil {
		website = *user.URL
	}

	return &Claims{
		Issuer:            endpoint,
		Subject:           user.IDStr,
		Name:              user.Name,
		Picture:           user.ProfileImageURLHTTPS,
		Email:             user.Email,
		PreferredUsername: user.ScreenName,
		Website:           website,
	}, nil
}

type xUser struct {
	ID                     int      `json:"id"`
	IDStr                  string   `json:"id_str"`
	Name                   string   `json:"name"`
	ScreenName             string   `json:"screen_name"`
	Location               string   `json:"location"`
	Description            string   `json:"description"`
	URL                    *string  `json:"url,omitempty"`
	Protected              bool     `json:"protected"`
	FollowersCount         int      `json:"followers_count"`
	FriendsCount           int      `json:"friends_count"`
	ListedCount            int      `json:"listed_count"`
	CreatedAt              string   `json:"created_at"`
	FavouritesCount        int      `json:"favourites_count"`
	Verified               bool     `json:"verified"`
	StatusesCount          int      `json:"statuses_count"`
	DefaultProfile         bool     `json:"default_profile"`
	DefaultProfileImage    bool     `json:"default_profile_image"`
	ProfileImageURLHTTPS   string   `json:"profile_image_url_https"`
	WithheldInCountries    []string `json:"withheld_in_countries"`
	Suspended              bool     `json:"suspended"`
	NeedsPhoneVerification bool     `json:"needs_phone_verification"`
	Email                  string   `json:"email"`
}
