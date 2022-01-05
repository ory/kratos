package oidc

import (
	"context"
	"net/http"

	"github.com/ory/kratos/cipher"
	"golang.org/x/oauth2"
)

type oAuth2Provider interface {
	Config() *Configuration
	OAuth2(context.Context) (*oauth2.Config, error)
	Claims(ctx context.Context, exchange *oauth2.Token) (*Claims, error)
}

type oAuth2Token struct {
	op     oAuth2Provider
	token  *oauth2.Token
	claims *Claims
}

func (t *oAuth2Token) Claims(ctx context.Context) (*Claims, error) {
	if t.claims == nil {
		c, err := t.op.Claims(ctx, t.token)
		if err != nil {
			return nil, err
		}

		t.claims = c
	}

	return t.claims, nil
}

func (t *oAuth2Token) CredentialsConfig(ctx context.Context, c cipher.Cipher) (*ProviderCredentialsConfig, error) {
	var (
		it  string
		err error
	)

	if idToken, ok := t.token.Extra("id_token").(string); ok {
		if it, err = c.Encrypt(ctx, []byte(idToken)); err != nil {
			return nil, err
		}
	}

	cat, err := c.Encrypt(ctx, []byte(t.token.AccessToken))
	if err != nil {
		return nil, err
	}

	crt, err := c.Encrypt(ctx, []byte(t.token.RefreshToken))
	if err != nil {
		return nil, err
	}

	claims, err := t.Claims(ctx)
	if err != nil {
		return nil, err
	}

	return &ProviderCredentialsConfig{
		Subject:             claims.Subject,
		InitialAccessToken:  cat,
		InitialRefreshToken: crt,
		InitialIDToken:      it,
	}, nil
}

func oAuth2CodeURL(ctx context.Context, state string, op oAuth2Provider, opts ...oauth2.AuthCodeOption) (string, error) {
	c, err := op.OAuth2(ctx)
	if err != nil {
		return "", err
	}

	return c.AuthCodeURL(state, opts...), nil
}

func parseOAuth2Token(ctx context.Context, op oAuth2Provider, req *http.Request) (Token, error) {
	c, err := op.OAuth2(ctx)
	if err != nil {
		return nil, err
	}

	code := req.URL.Query().Get("code")

	token, err := c.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return &oAuth2Token{
		op:    op,
		token: token,
	}, nil
}
