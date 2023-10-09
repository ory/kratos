// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package request

import (
	"encoding/json"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

type (
	noopAuthStrategy struct{}

	basicAuthStrategy struct {
		user     string
		password string
	}

	apiKeyStrategy struct {
		name  string
		value string
		in    string
	}
)

func newNoopAuthStrategy(_ json.RawMessage) (AuthStrategy, error) {
	return &noopAuthStrategy{}, nil
}

func (c *noopAuthStrategy) apply(_ *retryablehttp.Request) {}

func newBasicAuthStrategy(raw json.RawMessage) (AuthStrategy, error) {
	type config struct {
		User     string
		Password string
	}

	var c config
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}

	return &basicAuthStrategy{
		user:     c.User,
		password: c.Password,
	}, nil
}

func (c *basicAuthStrategy) apply(req *retryablehttp.Request) {
	req.SetBasicAuth(c.user, c.password)
}

func newApiKeyStrategy(raw json.RawMessage) (AuthStrategy, error) {
	type config struct {
		In    string
		Name  string
		Value string
	}

	var c config
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}

	return &apiKeyStrategy{
		in:    c.In,
		name:  c.Name,
		value: c.Value,
	}, nil
}

func (c *apiKeyStrategy) apply(req *retryablehttp.Request) {
	switch c.in {
	case "cookie":
		req.AddCookie(&http.Cookie{Name: c.name, Value: c.value})
	default:
		req.Header.Set(c.name, c.value)
	}
}
