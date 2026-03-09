// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ory/x/logrusx"
)

// tokenExchangeLoggingTransport logs OIDC token endpoint request and response for debugging (e.g. PayPal).
// Secrets are redacted: client_secret, code, code_verifier in request; access_token, refresh_token, id_token in response.
type tokenExchangeLoggingTransport struct {
	base       http.RoundTripper
	log        *logrusx.Logger
	exchangeID string
	providerID string
}

func (t *tokenExchangeLoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only log OAuth2 token endpoint calls (e.g. PayPal .../v1/oauth2/token)
	if req.URL == nil || !strings.Contains(req.URL.Path, "oauth2/token") {
		return t.base.RoundTrip(req)
	}

	// Log request (redact secrets)
	var reqBody string
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(body))
		reqBody = redactTokenRequestForm(string(body))
	}
	t.log.WithField("oidc_exchange_id", t.exchangeID).
		WithField("provider_id", t.providerID).
		WithField("request_url", req.URL.String()).
		WithField("request_method", req.Method).
		WithField("request_body_redacted", reqBody).
		Info("OIDC token endpoint request (secrets redacted)")

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Log what provider (e.g. PayPal) returned, straight away (tokens redacted).
	if resp.Body != nil {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(body))
		redacted := redactTokenResponseJSON(body)
		t.log.WithField("oidc_exchange_id", t.exchangeID).
			WithField("provider_id", t.providerID).
			WithField("response_status", resp.StatusCode).
			Infof("PayPal/token endpoint response: %s", redacted)
	}

	return resp, nil
}

// redactTokenRequestForm redacts client_secret, code, code_verifier in application/x-www-form-urlencoded body.
func redactTokenRequestForm(body string) string {
	vals, err := url.ParseQuery(body)
	if err != nil {
		return "[parse error]"
	}
	redact := []string{"client_secret", "code", "code_verifier"}
	for _, k := range redact {
		if vals.Has(k) {
			vals.Set(k, "[REDACTED]")
		}
	}
	return vals.Encode()
}

// redactTokenResponseJSON redacts access_token, refresh_token, id_token in JSON; shows keys and presence.
func redactTokenResponseJSON(body []byte) string {
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return string(body) // fallback: return as-is if not JSON
	}
	for _, k := range []string{"access_token", "refresh_token", "id_token"} {
		if _, ok := m[k]; ok {
			m[k] = "[REDACTED]"
		}
	}
	out, _ := json.Marshal(m)
	return string(out)
}
