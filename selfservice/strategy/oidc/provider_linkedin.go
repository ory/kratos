// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
)

type Profile struct {
	LocalizedLastName  string `json:"localizedLastName"`
	LocalizedFirstName string `json:"localizedFirstName"`
	ProfilePicture     *struct {
		DisplayImage struct {
			Elements []struct {
				Identifiers []struct {
					Identifier string `json:"identifier"`
				} `json:"identifiers"`
			} `json:"elements"`
		} `json:"displayImage~"`
	} `json:"profilePicture",omitempty`
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

func (l *ProviderLinkedIn) ProfilePicture(profile *Profile) string {
	if profile.ProfilePicture != nil {
		var elements = (*profile.ProfilePicture).DisplayImage.Elements
		switch len(elements) {
		case 0:
			return ""
		case 1:
			return elements[0].Identifiers[0].Identifier
		case 2:
			return elements[1].Identifiers[0].Identifier
		default:
			return elements[2].Identifiers[0].Identifier
		}
	}

	return ""
}

func (l *ProviderLinkedIn) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {

	var profile *Profile
	var emailaddress *EmailAddress
	var profilePicture string

	o, err := l.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := l.reg.HTTPClient(ctx, httpx.ResilientClientWithClient(o.Client(ctx, exchange)))
	profile, err = l.Profile(client)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	emailaddress, err = l.Email(client)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	profilePicture = l.ProfilePicture(profile)

	claims := &Claims{
		Subject:   profile.ID,
		Issuer:    "https://login.linkedin.com/",
		Email:     emailaddress.Elements[0].Handle.EmailAddress,
		GivenName: profile.LocalizedFirstName,
		LastName:  profile.LocalizedLastName,
		Picture:   profilePicture,
	}

	return claims, nil
}
