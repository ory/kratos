// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/jsonschema/v3/httploader"
	"github.com/ory/kratos/driver/config"
	. "github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/x/configx"
	"github.com/ory/x/httpx"
)

func TestSchemaValidatorDisallowsInternalNetworkRequests(t *testing.T) {
	t.Parallel()

	_, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.IdentitySchemasConfig(map[string]string{
			"localhost":  "https://localhost/schema/whatever",
			"privateRef": "file://stub/localhost-ref.schema.json",
			"inlineRef":  "base64://" + base64.StdEncoding.EncodeToString([]byte(`{"traits": {}}`)),
		})),
		configx.WithValue(config.ViperKeyClientHTTPNoPrivateIPRanges, true),
	)

	v := NewValidator(reg)

	for id, expectedErr := range map[string]string{
		"localhost":  "is not a permitted destination",
		"privateRef": "is not a permitted destination",
		"inlineRef":  "",
	} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()

			i := &Identity{
				SchemaID: id,
				Traits:   Traits(`{ "firstName": "first-name", "lastName": "last-name", "age": 1 }`),
			}
			ctx := context.WithValue(t.Context(), httploader.ContextKey, reg.HTTPClient(t.Context()))
			err := v.Validate(ctx, i)
			if expectedErr == "" {
				assert.NoError(t, err)
				return
			}

			var hErr *herodot.DefaultError
			require.ErrorAs(t, err, &hErr)
			assert.Contains(t, hErr.Debug(), expectedErr)
		})
	}
}

func TestSchemaValidator(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	router := http.NewServeMux()
	router.HandleFunc("GET /schema/{name}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
  "$id": "https://example.com/person.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Person",
  "type": "object",
  "properties": {
	"traits": {
	  "type": "object",
	  "properties": {
        "` + r.PathValue("name") + `": {
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

	_, reg := pkg.NewFastRegistryWithMocks(t,
		configx.WithValues(testhelpers.IdentitySchemasConfig(map[string]string{
			"default":         ts.URL + "/schema/firstName",
			"whatever":        ts.URL + "/schema/whatever",
			"unreachable-url": ts.URL + "/404-not-found",
		})),
	)
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
			err: "Invalid configuration",
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			ctx := context.WithValue(t.Context(), httploader.ContextKey, httpx.NewResilientClient())
			err := v.Validate(ctx, tc.i)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.err)
			}
		})
	}
}
