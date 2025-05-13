// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package request

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

type (
	noopAuthStrategy  struct{}
	basicAuthStrategy struct {
		user     string
		password string
	}
	apiKeyStrategy struct {
		name  string
		value string
		in    string
	}
	AuthStrategy interface {
		apply(req *retryablehttp.Request)
	}
)

func authStrategy(typ string, config map[string]any) (AuthStrategy, error) {
	switch typ {
	case "":
		return NewNoopAuthStrategy(), nil
	case "api_key":
		name, ok := config["name"].(string)
		if !ok {
			return nil, fmt.Errorf("api_key auth strategy requires a string name")
		}
		value, ok := config["value"].(string)
		if !ok {
			return nil, fmt.Errorf("api_key auth strategy requires a string value")
		}
		in, _ := config["in"].(string) // in is optional
		return NewAPIKeyStrategy(in, name, value), nil
	case "basic_auth":
		user, ok := config["user"].(string)
		if !ok {
			return nil, fmt.Errorf("basic_auth auth strategy requires a string user")
		}
		password, ok := config["password"].(string)
		if !ok {
			return nil, fmt.Errorf("basic_auth auth strategy requires a string password")
		}
		return NewBasicAuthStrategy(user, password), nil
	}

	return nil, fmt.Errorf("unsupported auth type: %s", typ)
}

func NewNoopAuthStrategy() AuthStrategy {
	return &noopAuthStrategy{}
}

func (c *noopAuthStrategy) apply(_ *retryablehttp.Request) {}

func NewBasicAuthStrategy(user, password string) AuthStrategy {
	return &basicAuthStrategy{
		user:     user,
		password: password,
	}
}

func (c *basicAuthStrategy) apply(req *retryablehttp.Request) {
	req.SetBasicAuth(c.user, c.password)
}

func NewAPIKeyStrategy(in, name, value string) AuthStrategy {
	return &apiKeyStrategy{
		in:    in,
		name:  name,
		value: value,
	}
}

func (c *apiKeyStrategy) apply(req *retryablehttp.Request) {
	switch c.in {
	case "cookie":
		req.AddCookie(&http.Cookie{Name: c.name, Value: c.value})
	default:
		// TODO add deprecation warning
		fallthrough
	case "header", "":
		req.Header.Set(c.name, c.value)
	}
}
