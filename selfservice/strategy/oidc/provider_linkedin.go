package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"

	"github.com/ory/herodot"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"
)

type Profile struct {
	LocalizedFirstName string `json:"localizedFirstName"`
	LocalizedLastName  string `json:"localizedLastName"`
}

type EmailAddress struct {
	Elements []struct {
		Handle struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"handle~"`
		HandleUrn string `json:"handle"`
	} `json:"elements"`
}

// type APIUrl string

const (
	ProfileUrl string = "https://api.linkedin.com/v2/me"
	EmailUrl   string = "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))"
)

type ProviderLinkedIn struct {
	config *Configuration
	reg    dependencies
}

func NewProviderLinkedIn(
	config *Configuration,
	reg dependencies,
) *ProviderLinkedIn {
	return &ProviderLinkedIn{
		config: config,
		reg:    reg,
	}
}

func (l *ProviderLinkedIn) Config() *Configuration {
	return l.config
}

func (l *ProviderLinkedIn) oauth2(ctx context.Context) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     l.config.ClientID,
		ClientSecret: l.config.ClientSecret,
		Endpoint:     linkedin.Endpoint,
		Scopes:       l.config.Scope,
		RedirectURL:  l.config.Redir(l.reg.Config().SelfPublicURL(ctx)),
	}
}

func (l *ProviderLinkedIn) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return l.oauth2(ctx), nil
}

func (l *ProviderLinkedIn) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (l *ProviderLinkedIn) ApiCall(url string, result interface{}, exchange *oauth2.Token) error {
	var bearer = "Bearer " + exchange.AccessToken
	req, err := http.NewRequest("GET", string(url), nil)
	if err != nil {
		return errors.WithStack(err)
	}
	req.Header.Add("Authorization", bearer)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.WithStack(err)
	}
	err = json.Unmarshal(body, result)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (l *ProviderLinkedIn) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	grantedScopes := stringsx.Splitx(fmt.Sprintf("%s", exchange.Extra("scope")), ",")
	for _, check := range l.Config().Scope {
		if !stringslice.Has(grantedScopes, check) {
			return nil, errors.WithStack(ErrScopeMissing)
		}
	}

	var profile Profile
	var emailaddress EmailAddress
	err := l.ApiCall(ProfileUrl, &profile, exchange)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	err = l.ApiCall(EmailUrl, &emailaddress, exchange)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	claims := &Claims{
		Email:    emailaddress.Elements[0].Handle.EmailAddress,
		Name:     profile.LocalizedFirstName,
		LastName: profile.LocalizedLastName,
	}

	return claims, nil
}
