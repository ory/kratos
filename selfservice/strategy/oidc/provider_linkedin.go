// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/ory/x/otelx"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
)

type LinkedInProfile struct {
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
	} `json:"profilePicture,omitempty"`
	ID string `json:"id"`
}

type LinkedInEmail struct {
	Elements []struct {
		Handle struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"handle~"`
		HandleUrn string `json:"handle"`
	} `json:"elements"`
}

type LinkedInIntrospection struct {
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
		RedirectURL:  l.config.Redir(l.reg.Config().OIDCRedirectURIBase(ctx)),
	}
}

func (l *ProviderLinkedIn) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return l.oauth2(ctx), nil
}

func (l *ProviderLinkedIn) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (l *ProviderLinkedIn) fetch(ctx context.Context, client *retryablehttp.Client, url string, result interface{}) (err error) {
	ctx, span := l.reg.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.oidc.ProviderLinkedIn.fetch")
	defer otelx.End(span, &err)

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	res, err := client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}

	defer res.Body.Close()
	if err := logUpstreamError(l.reg.Logger(), res); err != nil {
		return err
	}

	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (l *ProviderLinkedIn) Profile(ctx context.Context, client *retryablehttp.Client) (*LinkedInProfile, error) {
	var result LinkedInProfile
	if err := l.fetch(ctx, client, ProfileUrl, &result); err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

func (l *ProviderLinkedIn) Email(ctx context.Context, client *retryablehttp.Client) (*LinkedInEmail, error) {
	var result LinkedInEmail
	if err := l.fetch(ctx, client, EmailUrl, &result); err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

func (l *ProviderLinkedIn) ProfilePicture(profile *LinkedInProfile) string {
	if profile.ProfilePicture == nil {
		return ""
	}

	elements := profile.ProfilePicture.DisplayImage.Elements
	i := len(elements)
	if i == 0 {
		return ""
	} else if i > 3 {
		i = 3
	}

	identifiers := elements[i-1].Identifiers
	if len(identifiers) == 0 {
		return ""
	}

	return identifiers[0].Identifier
}

func (l *ProviderLinkedIn) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (_ *Claims, err error) {
	ctx, span := l.reg.Tracer(ctx).Tracer().Start(ctx, "selfservice.strategy.oidc.ProviderLinkedIn.Claims")
	defer otelx.End(span, &err)

	o, err := l.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	client := l.reg.HTTPClient(ctx, httpx.ResilientClientWithClient(o.Client(ctx, exchange)))
	profile, err := l.Profile(ctx, client)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	email, err := l.Email(ctx, client)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	claims := &Claims{
		Subject:   profile.ID,
		Issuer:    "https://login.linkedin.com/",
		Email:     email.Elements[0].Handle.EmailAddress,
		GivenName: profile.LocalizedFirstName,
		LastName:  profile.LocalizedLastName,
		Picture:   l.ProfilePicture(profile),
	}

	return claims, nil
}
