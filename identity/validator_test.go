// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/negroni"

	"github.com/ory/jsonschema/v3/httploader"
	"github.com/ory/kratos/x"
	"github.com/ory/x/httpx"

	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	. "github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
)

func TestSchemaValidatorDisallowsInternalNetworkRequests(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	conf.MustSet(ctx, config.ViperKeyClientHTTPNoPrivateIPRanges, true)
	conf.MustSet(ctx, config.ViperKeyIdentitySchemas, []config.Schema{
		{ID: "localhost", URL: "https://localhost/schema/whatever"},
		{ID: "privateRef", URL: "file://stub/localhost-ref.schema.json"},
	})

	v := NewValidator(reg)
	n := negroni.New(x.HTTPLoaderContextMiddleware(reg))
	router := httprouter.New()
	router.GET("/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		i := &Identity{
			SchemaID: ps.ByName("id"),
			Traits:   Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
		}
		_, _ = w.Write([]byte(fmt.Sprintf("%+v", v.Validate(r.Context(), i))))
	})
	n.UseHandler(router)

	ts := httptest.NewServer(n)
	t.Cleanup(ts.Close)

	// Make the request
	do := func(t *testing.T, id string) string {
		res, err := ts.Client().Get(ts.URL + "/" + id)
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		return string(body)
	}

	for _, tc := range [][2]string{
		{"localhost", "is not a public IP address"},
		{"privateRef", "is not a public IP address"},
	} {
		t.Run(fmt.Sprintf("case=%s", tc[0]), func(t *testing.T) {
			assert.Contains(t, do(t, tc[0]), tc[1])
		})
	}
}

func TestSchemaValidator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	router := httprouter.New()
	router.GET("/schema/:name", func(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/person.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
	"traits": {
	  "type": "object",
	  "properties": {
        "` + ps.ByName("name") + `": {
          "type": "string",
          "description": "The person's first name."
        },
        "lastName": {
          "type": "string",
          "description": "The person's last name."
        },
        "age": {
          "description": "Age in years which must be equal to or greater than zero.",
          "type": "integer",
          "minimum": 1
        }
	  },
	  "additionalProperties": false
	}
  },
  "additionalProperties": false
}`))
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	conf, reg := internal.NewFastRegistryWithMocks(t)
	conf.MustSet(ctx, config.ViperKeyIdentitySchemas, []config.Schema{
		{ID: "default", URL: ts.URL + "/schema/firstName"},
		{ID: "whatever", URL: ts.URL + "/schema/whatever"},
		{ID: "unreachable-url", URL: ts.URL + "/404-not-found"},
	})
	v := NewValidator(reg)

	for k, tc := range []struct {
		i   *Identity
		err string
	}{
		{
			i: &Identity{
				Traits: Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			},
		},
		{
			i: &Identity{
				Traits: Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": -1 }`),
			},
			err: "I[#/traits/age] S[#/properties/traits/properties/age/minimum] must be >= 1 but found -1",
		},
		{
			i: &Identity{
				Traits: Traits(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
			},
			err: `I[#/traits] S[#/properties/traits/additionalProperties] additionalProperties "whatever" not allowed`,
		},
		{
			i: &Identity{
				SchemaID: "whatever",
				Traits:   Traits(`{ "whatever": "first-name", "lastName": "last-name", "age": 1 }`),
			},
		},
		{
			i: &Identity{
				SchemaID: "whatever",
				Traits:   Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			},
			err: `I[#/traits] S[#/properties/traits/additionalProperties] additionalProperties "firstName" not allowed`,
		},
		{
			i: &Identity{
				SchemaID: "unreachable-url",
				Traits:   Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			},
			err: "An internal server error occurred, please contact the system administrator",
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			ctx := context.WithValue(ctx, httploader.ContextKey, httpx.NewResilientClient())
			err := v.Validate(ctx, tc.i)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.err)
			}
		})
	}
}
