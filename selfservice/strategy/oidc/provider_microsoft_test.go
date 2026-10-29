// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/x/configx"
)

// captureTransport records the token request and answers with a static token response.
type captureTransport struct {
	req  *http.Request
	form url.Values
}

func (c *captureTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	c.req = r
	c.form, err = url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"access_token":"at","token_type":"Bearer","expires_in":3600,"id_token":"idt"}`)),
		Request:    r,
	}, nil
}

func TestProviderMicrosoftExchange(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValue(config.ViperKeyPublicBaseURL, "https://ory.sh"))

	exchange := func(t *testing.T, c *oidc.Configuration) (*captureTransport, *oauth2.Token, error) {
		transport := &captureTransport{}
		// x/oauth2 reads the HTTP client from the context, same as Strategy.exchangeCode does.
		ctx := context.WithValue(t.Context(), oauth2.HTTPClient, &http.Client{Transport: transport})
		p := oidc.NewProviderMicrosoft(c, reg).(oidc.OAuth2TokenExchanger)
		token, err := p.Exchange(ctx, "the-code")
		return transport, token, err
	}

	base := func() *oidc.Configuration {
		return &oidc.Configuration{
			Provider: "microsoft",
			ID:       "microsoft",
			ClientID: "client-id",
			Tenant:   "organizations",
		}
	}

	t.Run("case=uses client_secret when no assertion file is configured", func(t *testing.T) {
		c := base()
		c.ClientSecret = "secret"

		tr, token, err := exchange(t, c)
		require.NoError(t, err)
		assert.Equal(t, "at", token.AccessToken)
		assert.Equal(t, "https://login.microsoftonline.com/organizations/oauth2/v2.0/token", tr.req.URL.String())
		_, basicSecret, hasBasicAuth := tr.req.BasicAuth()
		assert.True(t, (hasBasicAuth && basicSecret == "secret") || tr.form.Get("client_secret") == "secret",
			"client_secret must be sent either as HTTP Basic auth or as a form value")
		assert.Empty(t, tr.form.Get("client_assertion"))
		assert.Empty(t, tr.form.Get("client_assertion_type"))
	})

	t.Run("case=sends client_assertion read from file instead of client_secret", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token")
		require.NoError(t, os.WriteFile(path, []byte("header.payload.signature\n"), 0o600))
		c := base()
		c.ClientSecret = "must-not-be-sent"
		c.ClientAssertionFile = path

		tr, token, err := exchange(t, c)
		require.NoError(t, err)
		assert.Equal(t, "at", token.AccessToken)
		assert.Equal(t, "the-code", tr.form.Get("code"))
		assert.Equal(t, "authorization_code", tr.form.Get("grant_type"))
		assert.Equal(t, "client-id", tr.form.Get("client_id"))
		assert.Equal(t, "urn:ietf:params:oauth:client-assertion-type:jwt-bearer", tr.form.Get("client_assertion_type"))
		assert.Equal(t, "header.payload.signature", tr.form.Get("client_assertion"))
		_, hasClientSecret := tr.form["client_secret"]
		assert.False(t, hasClientSecret, "client_secret must not be sent when a client assertion is used")
		assert.Empty(t, tr.req.Header.Get("Authorization"))
	})

	t.Run("case=re-reads the assertion file on every exchange", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token")
		require.NoError(t, os.WriteFile(path, []byte("first"), 0o600))
		c := base()
		c.ClientAssertionFile = path

		tr, _, err := exchange(t, c)
		require.NoError(t, err)
		assert.Equal(t, "first", tr.form.Get("client_assertion"))

		require.NoError(t, os.WriteFile(path, []byte("second"), 0o600))
		tr, _, err = exchange(t, c)
		require.NoError(t, err)
		assert.Equal(t, "second", tr.form.Get("client_assertion"))
	})

	t.Run("case=fails with misconfiguration when assertion file is empty", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token")
		require.NoError(t, os.WriteFile(path, []byte("  \n"), 0o600))
		c := base()
		c.ClientAssertionFile = path

		tr, _, err := exchange(t, c)
		require.Error(t, err)
		assert.ErrorIs(t, err, herodot.ErrMisconfiguration())
		assert.Nil(t, tr.req, "no request must be sent to the token endpoint")
	})

	t.Run("case=fails with misconfiguration when assertion file is missing", func(t *testing.T) {
		c := base()
		c.ClientAssertionFile = filepath.Join(t.TempDir(), "does-not-exist")

		tr, _, err := exchange(t, c)
		require.Error(t, err)
		assert.ErrorIs(t, err, herodot.ErrMisconfiguration())
		assert.Nil(t, tr.req, "no request must be sent to the token endpoint")
	})
}
