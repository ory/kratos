// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"
)

type Profile struct {
	LocalizedLastName  string `json:"localizedLastName"`
	LocalizedFirstName string `json:"localizedFirstName"`
	ProfilePicture     struct {
		DisplayImage struct {
			Elements []struct {
				Identifiers []struct {
					Identifier string `json:"identifier"`
				} `json:"identifiers"`
			} `json:"elements"`
		} `json:"displayImage~"`
	} `json:"profilePicture"`
	ID string `json:"id"`
}

type EmailAddress struct {
	Elements []struct {
		Handle struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"handle~"`
		HandleUrn string `json:"handle"`
	} `json:"elements"`
}

type Introspection struct {
	Active       bool   `json:"active"`
	ClientID     string `json:"client_id"`
	AuthorizedAt uint32 `json:"authorized_at"`
	CreatedAt    uint32 `json:"created_at"`
	ExpiresAt    uint32 `json:"expires_at"`
	Status       string `json:"status"`
	Scope        string `json:"scope"`
	AuthType     string `json:"auth_type"`
}

// type APIUrl string

const (
	ProfileUrl       string = "https://api.linkedin.com/v2/me?projection=(id,localizedFirstName,localizedLastName,profilePicture(displayImage~digitalmediaAsset:playableStreams))"
	EmailUrl         string = "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))"
	IntrospectionURL string = "https://www.linkedin.com/oauth/v2/introspectToken"
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

func (l *ProviderLinkedIn) ApiGetCall(client *retryablehttp.Client, url string, result interface{}) error {
	req, err := retryablehttp.NewRequest(http.MethodGet, string(url), nil)
	if err != nil {
		return errors.WithStack(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (l *ProviderLinkedIn) Profile(client *retryablehttp.Client) (*Profile, error) {
	var profile Profile

	if err := l.ApiGetCall(client, ProfileUrl, &profile); err != nil {
		return nil, errors.WithStack(err)
	}
	return &profile, nil
}

func (l *ProviderLinkedIn) Email(client *retryablehttp.Client) (*EmailAddress, error) {
	var emailaddress EmailAddress

	if err := l.ApiGetCall(client, EmailUrl, &emailaddress); err != nil {
		return nil, errors.WithStack(err)
	}

	return &emailaddress, nil
}

func (l *ProviderLinkedIn) Introspection(client *retryablehttp.Client, result interface{}, exchange *oauth2.Token) error {
	form := url.Values{"client_id": {l.config.ClientID}, "client_secret": {l.config.ClientSecret}, "token": {exchange.AccessToken}}
	req, err := retryablehttp.NewRequest(http.MethodPost, string(IntrospectionURL), strings.NewReader(form.Encode()))
	if err != nil {
		return errors.WithStack(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.PostForm = form
	resp, err := client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (l *ProviderLinkedIn) VerifyScopes(client *retryablehttp.Client, exchange *oauth2.Token) error {
	var introspection Introspection

	if err := l.Introspection(client, &introspection, exchange); err != nil {
		return errors.WithStack(err)
	}
	grantedScopes := stringsx.Splitx(introspection.Scope, ",")
	for _, check := range l.Config().Scope {
		if !stringslice.Has(grantedScopes, check) {
			return errors.WithStack(ErrScopeMissing)
		}
	}

	return nil
}

func (l *ProviderLinkedIn) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {

	var profile *Profile
	var emailaddress *EmailAddress

	o, err := l.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := l.reg.HTTPClient(ctx, httpx.ResilientClientWithClient(o.Client(ctx, exchange)))
	if err = l.VerifyScopes(client, exchange); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	profile, err = l.Profile(client)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	emailaddress, err = l.Email(client)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	claims := &Claims{
		Email:     emailaddress.Elements[0].Handle.EmailAddress,
		Name:      fmt.Sprintf("%s %s", profile.LocalizedFirstName, profile.LocalizedLastName),
		GivenName: profile.LocalizedFirstName,
		LastName:  profile.LocalizedLastName,
		Picture:   profile.ProfilePicture.DisplayImage.Elements[2].Identifiers[0].Identifier,
	}

	return claims, nil
}
