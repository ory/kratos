// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
)

var _ Provider = (*ProviderTwitter)(nil)
var _ OAuth1Provider = (*ProviderTwitter)(nil)

const twitterUserInfoBase = "https://api.twitter.com/1.1/account/verify_credentials.json"
const twitterUserInfoWithEmail = twitterUserInfoBase + "?include_email=true"

type ProviderTwitter struct {
	config *Configuration
	reg    Dependencies
}

func (p *ProviderTwitter) Config() *Configuration {
	return p.config
}

func NewProviderTwitter(
	config *Configuration,
	reg Dependencies) Provider {
	return &ProviderTwitter{
		config: config,
		reg:    reg,
	}
}

func (p *ProviderTwitter) ExchangeToken(ctx context.Context, req *http.Request) (*oauth1.Token, error) {
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

func (p *ProviderTwitter) AuthURL(ctx context.Context, state string) (string, error) {
	c := p.OAuth1(ctx)

	// We need to cheat so that callback validates on return
	c.CallbackURL = c.CallbackURL + fmt.Sprintf("?state=%s&code=unused", state)

	requestToken, _, err := c.RequestToken()
	if err != nil {
		return "", err
	}

	authzURL, err := c.AuthorizationURL(requestToken)
	if err != nil {
		return "", err
	}

	return authzURL.String(), nil
}

func (p *ProviderTwitter) CheckError(ctx context.Context, r *http.Request) error {
	if r.URL.Query().Get("denied") == "" {
		return nil
	}

	return errors.WithStack(herodot.ErrBadRequest.WithReasonf(`Unable to sign in with Twitter because the user denied the request.`))
}

func (p *ProviderTwitter) OAuth1(ctx context.Context) *oauth1.Config {
	return &oauth1.Config{
		ConsumerKey:    p.config.ClientID,
		ConsumerSecret: p.config.ClientSecret,
		Endpoint:       twitter.AuthorizeEndpoint,
		CallbackURL:    p.config.Redir(p.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (p *ProviderTwitter) userInfoEndpoint() string {
	for _, scope := range p.config.Scope {
		if scope == "email" {
			return twitterUserInfoWithEmail
		}
	}

	return twitterUserInfoBase
}

func (p *ProviderTwitter) Claims(ctx context.Context, token *oauth1.Token) (*Claims, error) {
	ctx = context.WithValue(ctx, oauth1.HTTPClient, p.reg.HTTPClient(ctx).HTTPClient)

	c := p.OAuth1(ctx)
	client := c.Client(ctx, token)
	endpoint := p.userInfoEndpoint()

	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	if err := logUpstreamError(p.reg.Logger(), resp); err != nil {
		return nil, err
	}

	user := &twitterUser{}
	if err := json.NewDecoder(resp.Body).Decode(user); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
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

type twitterUser struct {
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