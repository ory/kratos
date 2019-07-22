package oidc

import (
	"context"
	"encoding/json"

	"golang.org/x/oauth2"
)

type Provider interface {
	Config() *Configuration
	OAuth2(ctx context.Context) (*oauth2.Config, error)
	Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error)
}

type Claims struct {
	Subject  string
	Traits   json.RawMessage
	Metadata map[string]interface{}
}
