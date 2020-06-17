package oidc

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	glapi "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

var _ = glapi.ProjectClustersService{}

var endpoint = oauth2.Endpoint{
	TokenURL: "https://gitlab.com/oauth/token",
	AuthURL:  "https://gitlab.com/oauth/authorize",
}

type ProviderGitLab struct {
	config *Configuration
	public *url.URL
}

func NewProviderGitLab(
	config *Configuration,
	public *url.URL,
) *ProviderGitLab {
	return &ProviderGitLab{config: config, public: public}
}

func (g *ProviderGitLab) Config() *Configuration {
	return g.config
}

func (g *ProviderGitLab) oauth2() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint:     endpoint,
		Scopes:       g.config.Scope,
		RedirectURL:  g.config.Redir(g.public),
	}
}

func (g *ProviderGitLab) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(), nil
}

func (g *ProviderGitLab) AuthCodeURLOptions(r request) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (g *ProviderGitLab) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	tokenSource := oauth2.StaticTokenSource(exchange)
	client := oauth2.NewClient(ctx, tokenSource)
	req, err := http.NewRequest("GET", "https://gitlab.com/oauth/userinfo", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var claims Claims
	err = json.Unmarshal(body, &claims)
	if err != nil {
		return nil, err
	}
	return &claims, nil
}
