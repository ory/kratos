package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
	"github.com/ory/herodot"
	"github.com/pkg/errors"
)

var _ Provider = (*ProviderTwitter)(nil)

const twitterUserInfoEndpoint = "https://api.twitter.com/1.1/account/verify_credentials.json?include_email=true"

type ProviderTwitter struct {
	config *Configuration
	public *url.URL
}

func NewProviderTwitter(config *Configuration, public *url.URL) *ProviderTwitter {
	return &ProviderTwitter{
		config: config,
		public: public,
	}
}

func (p *ProviderTwitter) Token(ctx context.Context, req *http.Request) (Token, error) {
	requestToken, verifier, err := oauth1.ParseAuthorizationCallback(req)
	if err != nil {
		return nil, err
	}

	accessToken, accessSecret, err := p.OAuth1(ctx).AccessToken(requestToken, "", verifier)
	if err != nil {
		return nil, err
	}

	token := oauth1.NewToken(accessToken, accessSecret)

	return &oAuth1Token{token: token, op: p}, nil
}

func (p *ProviderTwitter) RedirectURL(ctx context.Context, state string, _ ider) (string, error) {
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

func (p *ProviderTwitter) Config() *Configuration {
	return p.config
}

func (p *ProviderTwitter) OAuth1(ctx context.Context) *oauth1.Config {
	return &oauth1.Config{
		ConsumerKey:    p.config.ClientID,
		ConsumerSecret: p.config.ClientSecret,
		Endpoint:       twitter.AuthorizeEndpoint,
		CallbackURL:    p.config.Redir(p.public),
	}
}

func (p *ProviderTwitter) Claims(ctx context.Context, token *oauth1.Token) (*Claims, error) {

	c := p.OAuth1(ctx)

	client := c.Client(ctx, token)

	resp, err := client.Get(twitterUserInfoEndpoint)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	defer resp.Body.Close()

	user := &twitterUser{}

	if err := json.NewDecoder(resp.Body).Decode(user); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	website := ""
	if user.URL != nil {
		website = *user.URL
	}

	return &Claims{
		Issuer:            twitterUserInfoEndpoint,
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
