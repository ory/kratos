package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

type ProviderFacebook struct {
	*ProviderGenericOIDC
}

func NewProviderFacebook(
	config *Configuration,
	public *url.URL,
) *ProviderFacebook {
	config.IssuerURL = "https://facebook.com"
	return &ProviderFacebook{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			public: public,
		},
	}
}

func (g *ProviderFacebook) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	p, err := g.provider(ctx)
	if err != nil {
		return nil, err
	}

	endpoint := p.Endpoint()

	if endpoint.AuthURL == "" {
		endpoint.AuthURL = "https://facebook.com/dialog/oauth"
	}
	if endpoint.TokenURL == "" {
		endpoint.TokenURL = "https://graph.facebook.com/oauth/access_token"
	}

	return g.oauth2ConfigFromEndpoint(endpoint), nil
}

func (g *ProviderFacebook) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := o.Client(ctx, exchange)

	u, err := url.Parse("https://graph.facebook.com/me?fields=id,name,first_name,last_name,middle_name,email,picture,birthday,gender")
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	var user struct {
		Id            string `json:"id,omitempty"`
		Name          string `json:"name,omitempty"`
		FirstName     string `json:"first_name,omitempty"`
		LastName      string `json:"last_name,omitempty"`
		MiddleName    string `json:"middle_name,omitempty"`
		Email         string `json:"email,omitempty"`
		EmailVerified bool
		Picture       struct {
			Data struct {
				Url string `json:"url,omitempty"`
			} `json:"data,omitempty"`
		} `json:"picture,omitempty"`
		BirthDay string `json:"birthday,omitempty"`
		Gender   string `json:"gender,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	if user.Email != "" {
		user.EmailVerified = true
	}

	return &Claims{
		Issuer:            u.String(),
		Subject:           user.Id,
		Name:              user.Name,
		GivenName:         user.FirstName,
		FamilyName:        user.LastName,
		MiddleName:        user.MiddleName,
		Nickname:          user.Name,
		PreferredUsername: user.Name,
		Picture:           user.Picture.Data.Url,
		Email:             user.Email,
		EmailVerified:     user.EmailVerified,
		Gender:            user.Gender,
		Birthdate:         user.BirthDay,
	}, nil
}
