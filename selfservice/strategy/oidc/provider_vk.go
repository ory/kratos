package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

type ProviderVK struct {
	config *Configuration
	public *url.URL
}

func NewProviderVK(
	config *Configuration,
	public *url.URL,
) *ProviderVK {
	return &ProviderVK{
		config: config,
		public: public,
	}
}

func (g *ProviderVK) Config() *Configuration {
	return g.config
}

func (g *ProviderVK) oauth2() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.vk.com/authorize",
			TokenURL: "https://oauth.vk.com/access_token",
		},
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.public),
	}
}

func (g *ProviderVK) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (g *ProviderVK) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(), nil
}

func (g *ProviderVK) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {

	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := o.Client(ctx, exchange)

	u, err := url.Parse("https://api.vk.com/method/users.get?fields=photo_200,nickname,bdate,sex&access_token=" + exchange.AccessToken + "&v=5.103")
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

	type User struct {
		Id            int    `json:"id,omitempty"`
		FirstName     string `json:"first_name,omitempty"`
		LastName      string `json:"last_name,omitempty"`
		Nickname      string `json:"nickname,omitempty"`
		Picture       string `json:"photo_200,omitempty"`
		Email         string
		EmailVerified bool
		Gender        int    `json:"sex,omitempty"`
		BirthDay      string `json:"bdate,omitempty"`
	}

	var response struct {
		Result []User `json:"response,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	user := response.Result[0]

	if email := exchange.Extra("email"); email != nil {
		user.Email = email.(string)
		user.EmailVerified = true
	}

	gender := ""
	switch user.Gender {
	case 1:
		gender = "female"
	case 2:
		gender = "male"
	}

	return &Claims{
		Issuer:        u.String(),
		Subject:       strconv.Itoa(user.Id),
		GivenName:     user.FirstName,
		FamilyName:    user.LastName,
		Nickname:      user.Nickname,
		Picture:       user.Picture,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Gender:        gender,
		Birthdate:     user.BirthDay,
	}, nil
}
