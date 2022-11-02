package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

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
	AuthorizedAt int32  `json:"authorized_at"`
	CreatedAt    int32  `json:"created_at"`
	ExpiresAt    int32  `json:"expires_at"`
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
	resp, err := http.PostForm(string(IntrospectionURL),
		url.Values{"client_id": {l.config.ClientID}, "client_secret": {l.config.ClientSecret}, "token": {exchange.AccessToken}})

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

func RedirectPost(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}

	lastReq := via[len(via)-1]
	if req.Response.StatusCode >= 300 && req.Response.StatusCode < 400 && lastReq.Method == http.MethodPost {
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

func LogStringToFile(message string) {
	// open file and create if non-existent
	file, err := os.OpenFile("kratos-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := log.New(file, "Custom Log ", log.LstdFlags)
	logger.Println(message)
}

func LogJsonToFile(objectName string, jsonObject interface{}) {
	introspectionJson, err := json.MarshalIndent(jsonObject, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	LogStringToFile(objectName + ": " + string(introspectionJson))
}

func (l *ProviderLinkedIn) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {

	var introspection Introspection
	var profile Profile
	var emailaddress EmailAddress

	LogStringToFile("Access token: " + exchange.AccessToken)

	err := l.Introspection(&introspection, exchange)
	LogJsonToFile("Introspection", introspection)
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
		Picture:  profile.ProfilePicture.DisplayImage.Elements[1].Identifiers[0].Identifier,
	}

	LogJsonToFile("Claims", claims)
	return claims, nil
}
