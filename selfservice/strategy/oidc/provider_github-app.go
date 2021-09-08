package oidc

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	ghapi "github.com/google/go-github/v27/github"

	"github.com/ory/herodot"
)

type ProviderGitHubApp struct {
	config *Configuration
	public *url.URL
}

func NewProviderGitHubApp(
	config *Configuration,
	public *url.URL,
) *ProviderGitHubApp {
	return &ProviderGitHubApp{
		config: config,
		public: public,
	}
}

func (g *ProviderGitHubApp) Config() *Configuration {
	return g.config
}

func (g *ProviderGitHubApp) oauth2() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.config.ClientID,
		ClientSecret: g.config.ClientSecret,
		Endpoint:     github.Endpoint,
		Scopes:       g.config.Scope,
		RedirectURL:  g.config.Redir(g.public),
	}
}

func (g *ProviderGitHubApp) OAuth2(ctx context.Context) (*oauth2.Config, error) {
	return g.oauth2(), nil
}

func (g *ProviderGitHubApp) AuthCodeURLOptions(r ider) []oauth2.AuthCodeOption {
	return []oauth2.AuthCodeOption{}
}

func (g *ProviderGitHubApp) Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error) {
	gh := ghapi.NewClient(g.oauth2().Client(ctx, exchange))

	user, _, err := gh.Users.Get(ctx, "")
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	claims := &Claims{
		Subject:   fmt.Sprintf("%d", user.GetID()),
		Issuer:    github.Endpoint.TokenURL,
		Name:      user.GetName(),
		Nickname:  user.GetLogin(),
		Website:   user.GetBlog(),
		Picture:   user.GetAvatarURL(),
		Profile:   user.GetHTMLURL(),
		UpdatedAt: user.GetUpdatedAt().Unix(),
	}

	// GitHub does not provide the user's private emails in the call to `/user`. Therefore, if scope "user:email" is set,
	// we want to make another request to `/user/emails` and merge that with our claims.
	emails, _, err := gh.Users.ListEmails(ctx, nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("%s", err))
	}

	for k, e := range emails {
		// If it is the primary email or it's the last email (no primary email set?), set the email.
		if e.GetPrimary() || k == len(emails) {
			claims.Email = e.GetEmail()
			claims.EmailVerified = e.GetVerified()
			break
		}
	}

	return claims, nil
}
