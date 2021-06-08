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

	// Force updatedAt to be an int if given as a string in the response.
	if updatedAtField := gjson.GetBytes(b, "updated_at"); updatedAtField.Exists() {
		v := updatedAtField.Value()
		switch v.(type) {
		case string:
			t, err := time.Parse(time.RFC3339, updatedAtField.String())
			updatedAt := t.Unix()

			// Unmarshal into generic map, replace the updated_at value with the correct type, then re-marshal.
			var data map[string]interface{}
			err = json.Unmarshal(b, &data)
			if err != nil {
				return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("bad type in response"))
			}

			// convert the correct int64 type back to a string, so we can Marshal it.
			data["updated_at"] = strconv.FormatInt(updatedAt, 10)

			// now remarshal so the unmarshal into Claims works.
			b, err = json.Marshal(data)
			if err != nil {
				return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
			}

		case float64:
			// nothing to do
			break

		default:
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("bad updated_at type"))
		}
	}

	// Once we get here, we know that if there is an updated_at field in the json, it is the correct type.
	var claims Claims
	if err := json.Unmarshal(b, &claims); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	return &claims, nil
}
