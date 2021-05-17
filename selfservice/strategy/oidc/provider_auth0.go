package oidc

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

type ProviderAuth0 struct {
	*ProviderGenericOIDC
}

func NewProviderAuth0(
	config *Configuration,
	public *url.URL,
) *ProviderAuth0 {
	return &ProviderAuth0{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			public: public,
		},
	}
}

func (g *ProviderAuth0) oauth2() (*oauth2.Config, error) {
	endpoint, err := url.Parse(g.config.IssuerURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	authUrl := *endpoint
	tokenUrl := *endpoint

	authUrl.Path = path.Join(authUrl.Path, "/authorize")
	tokenUrl.Path = path.Join(tokenUrl.Path, "/oauth/token")

	c := &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authUrl.String(),
			TokenURL: tokenUrl.String(),
		},
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.public),
	}

	return c, nil
}

func (g *ProviderAuth0) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2()
}

func (g *ProviderAuth0) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	o, err := g.OAuth2(ctx)

	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := o.Client(ctx, exchange)

	u, err := url.Parse(g.config.IssuerURL)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	u.Path = path.Join(u.Path, "/userinfo")
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	req.Header.Add("Authorization", "Bearer: "+exchange.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	// There is a bug in the response from Auth0. The updated_at field may be a string and not an int64.
	// https://community.auth0.com/t/oidc-id-token-claim-updated-at-violates-oidc-specification-breaks-rp-implementations/24098
	// We work around this by reading the json generically (as map[string]inteface{} and looking at the updated_at field
	// if it exists. If it's the wrong type (string), we fill out the claims by hand.

	// Once auth0 fixes this bug, all this workaround can be removed.
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	var values map[string]interface{}
	if err := json.Unmarshal(b, &values); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	if v, ok := values["updated_at"]; ok {
		// There is an updated_at field. Look at the type. If float64 (default for json numbers, then
		// we can just use the Claims. Otherwise we convert the value.
		if _, ok := v.(string); ok {
			// This is the bug. We need to convert.
			c, err := auth0Claims(b)
			if err != nil {
				return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
			}

			return c, nil
		}
	}

	// If we get here, we the claims are standard and we proceed normally.
	var claims Claims
	if err := json.Unmarshal(b, &claims); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	return &claims, nil
}

// This is not great.
func auth0Claims(b []byte) (*Claims, error) {

	type Auth0Claims struct {
		Issuer              string `json:"iss,omitempty"`
		Subject             string `json:"sub,omitempty"`
		Name                string `json:"name,omitempty"`
		GivenName           string `json:"given_name,omitempty"`
		FamilyName          string `json:"family_name,omitempty"`
		LastName            string `json:"last_name,omitempty"`
		MiddleName          string `json:"middle_name,omitempty"`
		Nickname            string `json:"nickname,omitempty"`
		PreferredUsername   string `json:"preferred_username,omitempty"`
		Profile             string `json:"profile,omitempty"`
		Picture             string `json:"picture,omitempty"`
		Website             string `json:"website,omitempty"`
		Email               string `json:"email,omitempty"`
		EmailVerified       bool   `json:"email_verified,omitempty"`
		Gender              string `json:"gender,omitempty"`
		Birthdate           string `json:"birthdate,omitempty"`
		Zoneinfo            string `json:"zoneinfo,omitempty"`
		Locale              string `json:"locale,omitempty"`
		PhoneNumber         string `json:"phone_number,omitempty"`
		PhoneNumberVerified bool   `json:"phone_number_verified,omitempty"`
		UpdatedAt           string `json:"updated_at,omitempty"`
	}
	var a Auth0Claims

	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}

	c := &Claims{}
	t, err := time.Parse(time.RFC3339, a.UpdatedAt) // apparently this is RFC3339
	if err == nil {
		c.UpdatedAt = t.Unix()
	} else {
		// Don't know what to do here. Just zero I guess.
		c.UpdatedAt = 0
	}

	c.Issuer = a.Issuer
	c.Subject = a.Subject
	c.Name = a.Name
	c.GivenName = a.GivenName
	c.FamilyName = a.FamilyName
	c.LastName = a.LastName
	c.MiddleName = a.MiddleName
	c.Nickname = a.Nickname
	c.PreferredUsername = a.PreferredUsername
	c.Profile = a.Profile
	c.Picture = a.Picture
	c.Website = a.Website
	c.Email = a.Email
	c.EmailVerified = a.EmailVerified
	c.Gender = a.Gender
	c.Birthdate = a.Birthdate
	c.Zoneinfo = a.Zoneinfo
	c.Locale = a.Locale
	c.PhoneNumber = a.PhoneNumber
	c.PhoneNumberVerified = a.PhoneNumberVerified

	return c, nil
}
