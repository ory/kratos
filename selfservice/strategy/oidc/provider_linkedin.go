package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"

	"github.com/ory/herodot"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"
)

type Profile struct {
	LocalizedLastName  string `json:"localizedLastName"`
	LocalizedFirstName string `json:"localizedFirstName"`
	ProfilePicture     struct {
		DisplayImage string `json:"displayImage"`
	} `json:"profilePicture"`
	FirstName struct {
		Localized struct {
			EnUS string `json:"en_US"`
		} `json:"localized"`
		PreferredLocale struct {
			Country  string `json:"country"`
			Language string `json:"language"`
		} `json:"preferredLocale"`
	} `json:"firstName"`
	LastName struct {
		Localized struct {
			EnUS string `json:"en_US"`
		} `json:"localized"`
		PreferredLocale struct {
			Country  string `json:"country"`
			Language string `json:"language"`
		} `json:"preferredLocale"`
	} `json:"lastName"`
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
	AuthorizedAt int32  `json:"authorized_at"`
	CreatedAt    int32  `json:"created_at"`
	ExpiresAt    int32  `json:"expires_at"`
	Status       string `json:"status"`
	Scope        string `json:"scope"`
	AuthType     string `json:"auth_type"`
}

// type APIUrl string

const (
	ProfileUrl       string = "https://api.linkedin.com/v2/me"
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

func (l *ProviderLinkedIn) ApiCall(url string, result interface{}, exchange *oauth2.Token) error {
	var bearer = "Bearer " + exchange.AccessToken
	req, err := http.NewRequest(http.MethodGet, string(url), nil)
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

func (l *ProviderLinkedIn) Introspection(result interface{}, exchange *oauth2.Token) error {
	v := url.Values{}
	v.Set("client_id", l.config.ClientID)
	v.Set("client_secret", l.config.ClientSecret)
	v.Set("token", exchange.AccessToken)
	req, err := http.NewRequest(http.MethodPost, string(IntrospectionURL), strings.NewReader(v.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return errors.WithStack(err)
	}
	client := &http.Client{
		CheckRedirect: redirectPostOn302,
	}
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

func redirectPostOn302(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}

	lastReq := via[len(via)-1]
	if req.Response.StatusCode == 302 && lastReq.Method == http.MethodPost {
		req.Method = http.MethodPost

		// Get the body of the original request, set here, since req.Body will be nil if a 302 was returned
		if via[0].GetBody != nil {
			var err error
			req.Body, err = via[0].GetBody()
			if err != nil {
				return err
			}
			req.ContentLength = via[0].ContentLength
		}
	}

	return nil
}

func (l *ProviderLinkedIn) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {

	var introspection Introspection
	var profile Profile
	var emailaddress EmailAddress

	err := l.Introspection(&introspection, exchange)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	grantedScopes := stringsx.Splitx(fmt.Sprintf("%s", introspection.Scope), ",")
	for _, check := range l.Config().Scope {
		if !stringslice.Has(grantedScopes, check) {
			return nil, errors.WithStack(ErrScopeMissing)
		}
	}
	err = l.ApiCall(ProfileUrl, &profile, exchange)
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
		Picture:  profile.ProfilePicture.DisplayImage,
	}

	return claims, nil
}
