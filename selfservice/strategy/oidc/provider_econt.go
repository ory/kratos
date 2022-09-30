package oidc

import (
	"context"
	"encoding/json"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/x/httpx"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
)

type ProviderEcont struct {
	*ProviderGenericOIDC
}

func NewProviderEcont(
	config *Configuration,
	reg dependencies,
) *ProviderEcont {
	return &ProviderEcont{
		ProviderGenericOIDC: &ProviderGenericOIDC{
			config: config,
			reg:    reg,
		},
	}
}

func (g *ProviderEcont) oauth2(ctx context.Context) (*oauth2.Config, error) {
	endpoint, err := g.endpoint()
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	authUrl := *endpoint
	tokenUrl := *endpoint

	authUrl.Path = path.Join(authUrl.Path, "/oauth2/auth")
	tokenUrl.Path = path.Join(tokenUrl.Path, "/oauth2/token")

	// Here we get the clientId & clientSecret from the kratos.yml file
	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authUrl.String(),
			TokenURL: tokenUrl.String(),
		},
		Scopes:      g.config.Scope,
		RedirectURL: g.config.Redir(g.reg.Config().OIDCRedirectURIBase(ctx)),
	}, nil
}

func (g *ProviderEcont) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(ctx)
}

func (g *ProviderEcont) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	conf, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	pTokenParams := &struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
		GrantType    string `json:"grant_type"`
	}{conf.ClientID, conf.ClientSecret, code, "authorization_code"}
	bs, err := json.Marshal(pTokenParams)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	r := strings.NewReader(string(bs))
	client := g.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs())

	req, err := retryablehttp.NewRequest("POST", conf.Endpoint.TokenURL, r)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	req.Header.Add("Content-Type", "application/json;charset=UTF-8")

	q := req.URL.Query()

	q.Add("grant_type", "authorization_code")
	q.Add("code", code)
	q.Add("client_id", conf.ClientID)
	q.Add("client_secret", conf.ClientSecret)

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	var dToken struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int32  `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&dToken); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	token := &oauth2.Token{
		AccessToken:  dToken.AccessToken,
		TokenType:    dToken.TokenType,
		RefreshToken: dToken.RefreshToken,
		Expiry:       time.Unix(time.Now().Unix()+int64(dToken.ExpiresIn), 0),
	}
	return token, nil
}

func (g *ProviderEcont) Claims(ctx context.Context, exchange *oauth2.Token, query url.Values) (*Claims, error) {
	o, err := g.OAuth2(ctx)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	client := g.reg.HTTPClient(ctx, httpx.ResilientClientDisallowInternalIPs(), httpx.ResilientClientWithClient(o.Client(ctx, exchange)))

	u, err := g.endpoint()
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	u.Path = path.Join(u.Path, "/oauth2/profile")

	req, err := retryablehttp.NewRequest("GET", u.String(), nil)

	q := req.URL.Query()
	q.Add("access_token", exchange.AccessToken)
	req.URL.RawQuery = q.Encode()

	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}
	defer resp.Body.Close()

	var claims Claims // <-- For additional claims we need to add them in this type! Not sure yet how for nested types.
	// Here we have decoded all the claims from econt (user data)
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	return &claims, nil
}

func (g *ProviderEcont) endpoint() (*url.URL, error) {
	var e = "https://login.econt.com" // defaultEndpoint
	if len(g.config.IssuerURL) > 0 {
		e = g.config.IssuerURL
	}
	return url.Parse(e)
}
