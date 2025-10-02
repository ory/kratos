// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package schema_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/client-go"
	_ "github.com/ory/jsonschema/v3/fileloader"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
	"github.com/ory/x/configx"
	"github.com/ory/x/contextx"
	"github.com/ory/x/urlx"
)

func TestHandler(t *testing.T) {
	router := x.NewTestRouterPublic(t)
	ts := contextx.NewConfigurableTestServer(router)
	t.Cleanup(ts.Close)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /identity.schema.json", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open("./stub/identity.schema.json")
		require.NoError(t, err)
		_, err = io.Copy(w, file)
		require.NoError(t, err)
	})
	mux.HandleFunc("GET /500", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})
	fileServer := httptest.NewServer(mux)
	t.Cleanup(fileServer.Close)

	schemas := map[string]struct {
		uri                      string
		getRaw                   func() ([]byte, error)
		expectedHttpResponseCode int
	}{
		"default": {
			uri:                      "file://./stub/identity.schema.json",
			getRaw:                   func() ([]byte, error) { return os.ReadFile("./stub/identity.schema.json") },
			expectedHttpResponseCode: http.StatusOK,
		},
		"identity2": {
			uri:                      "file://./stub/identity-2.schema.json",
			getRaw:                   func() ([]byte, error) { return os.ReadFile("./stub/identity-2.schema.json") },
			expectedHttpResponseCode: http.StatusOK,
		},
		"base64": {
			uri: "base64://ewogICIkc2NoZW1hIjogImh0dHA6Ly9qc29uLXNjaGVtYS5vcmcvZHJhZnQtMDcvc2NoZW1hIyIsCiAgInR5cGUiOiAib2JqZWN0IiwKICAicHJvcGVydGllcyI6IHsKICAgICJiYXIiOiB7CiAgICAgICJ0eXBlIjogInN0cmluZyIKICAgIH0KICB9LAogICJyZXF1aXJlZCI6IFsKICAgICJiYXIiCiAgXQp9",
			getRaw: func() ([]byte, error) {
				return base64.StdEncoding.DecodeString("ewogICIkc2NoZW1hIjogImh0dHA6Ly9qc29uLXNjaGVtYS5vcmcvZHJhZnQtMDcvc2NoZW1hIyIsCiAgInR5cGUiOiAib2JqZWN0IiwKICAicHJvcGVydGllcyI6IHsKICAgICJiYXIiOiB7CiAgICAgICJ0eXBlIjogInN0cmluZyIKICAgIH0KICB9LAogICJyZXF1aXJlZCI6IFsKICAgICJiYXIiCiAgXQp9")
			},
			expectedHttpResponseCode: http.StatusOK,
		},
		"unreachable": {
			uri: "http://127.0.0.1:12345/unreachable-schema",
			getRaw: func() ([]byte, error) {
				return nil, fmt.Errorf("connection refused")
			},
			expectedHttpResponseCode: http.StatusBadGateway,
		},
		"no-file": {
			uri:                      "file://./stub/does-not-exist.schema.json",
			getRaw:                   func() ([]byte, error) { return nil, fmt.Errorf("no such file or directory") },
			expectedHttpResponseCode: http.StatusInternalServerError,
		},
		"directory": {
			uri:    "file://./stub",
			getRaw: func() ([]byte, error) { return nil, fmt.Errorf("is a directory") },
			// On an existing directory, `open(2)` succeeds but `read(2)` fails so it looks like an I/O error.
			expectedHttpResponseCode: http.StatusInternalServerError,
		},
		"file-network": {
			uri:                      fileServer.URL + "/identity.schema.json",
			getRaw:                   func() ([]byte, error) { return os.ReadFile("./stub/identity.schema.json") },
			expectedHttpResponseCode: http.StatusOK,
		},
		"file-network-not-found": {
			uri:                      fileServer.URL + "/not-found",
			getRaw:                   func() ([]byte, error) { return nil, fmt.Errorf("could not be found") },
			expectedHttpResponseCode: http.StatusInternalServerError,
		},
		"file-network-500": {
			uri:                      fileServer.URL + "/500",
			getRaw:                   func() ([]byte, error) { return nil, fmt.Errorf("giving up") },
			expectedHttpResponseCode: http.StatusBadGateway,
		},
		"preset://email": {
			uri:                      "file://./stub/identity-2.schema.json",
			getRaw:                   func() ([]byte, error) { return os.ReadFile("./stub/identity-2.schema.json") },
			expectedHttpResponseCode: http.StatusOK,
		},
	}
	configSchemas := make(config.Schemas, 0, len(schemas))
	for id, s := range schemas {
		configSchemas = append(configSchemas, config.Schema{
			ID:  id,
			URL: s.uri,
		})
	}

	_, reg := internal.NewFastRegistryWithMocks(t, configx.WithValues(map[string]any{
		config.ViperKeyPublicBaseURL:           ts.URL,
		config.ViperKeyDefaultIdentitySchemaID: "default",
		config.ViperKeyIdentitySchemas:         configSchemas,
	}))
	reg.SchemaHandler().RegisterPublicRoutes(router)

	getReq := func(ctx context.Context, t *testing.T, path string, expectCode int) []byte {
		res, err := ts.Client(ctx).Get(ts.URL + path)
		require.NoError(t, err)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		require.EqualValuesf(t, expectCode, res.StatusCode, "%s", body)
		return body
	}

	for id, s := range schemas {
		t.Run(fmt.Sprintf("case=get %s schema", id), func(t *testing.T) {
			t.Parallel()

			_, err := s.getRaw()
			actual := getReq(t.Context(), t, fmt.Sprintf("/schemas/%s", url.PathEscape(id)), s.expectedHttpResponseCode)
			require.True(t, json.Valid(actual), string(actual))

			switch s.expectedHttpResponseCode {
			case http.StatusOK:
				require.NoError(t, err)
			case http.StatusInternalServerError, http.StatusBadGateway:

			default:
				panic("unreachable")
			}
		})
	}

	t.Run("case=get schema with base64 encoded ID", func(t *testing.T) {
		t.Parallel()

		expected, err := schemas["preset://email"].getRaw()
		require.NoError(t, err)

		actual := getReq(t.Context(), t, "/schemas/"+base64.RawURLEncoding.EncodeToString([]byte("preset://email")), http.StatusOK)
		require.JSONEq(t, string(expected), string(actual))
	})

	t.Run("case=get all schemas", func(t *testing.T) {
		t.Parallel()

		defaultSchema, err := configSchemas.FindSchemaByID("default")
		require.NoError(t, err)
		identity2Schema, err := configSchemas.FindSchemaByID("identity2")
		require.NoError(t, err)
		ctx := contextx.WithConfigValue(t.Context(), config.ViperKeyIdentitySchemas, config.Schemas{*defaultSchema, *identity2Schema})

		getSchemasPaginated := func(t *testing.T, page, perPage, expectCode int) []byte {
			return getReq(ctx, t, fmt.Sprintf("/schemas?page=%d&per_page=%d", page, perPage), expectCode)
		}

		body := getSchemasPaginated(t, 0, 10, http.StatusOK)

		var result []client.IdentitySchemaContainer
		require.NoErrorf(t, json.Unmarshal(body, &result), "%s", body)

		var actualIDs []string
		for _, s := range result {
			actualIDs = append(actualIDs, s.Id)
		}
		assert.Equal(t, []string{defaultSchema.ID, identity2Schema.ID}, actualIDs)

		assertCorrectSchema := func(t *testing.T, r client.IdentitySchemaContainer) {
			expected, err := schemas[r.Id].getRaw()
			require.NoError(t, err)
			actual, err := json.Marshal(r.Schema)
			require.NoError(t, err)
			assert.JSONEq(t, string(expected), string(actual))
		}

		for _, r := range result {
			assertCorrectSchema(t, r)
		}

		for page := range 2 {
			t.Run(fmt.Sprintf("page=%d", page), func(t *testing.T) {
				body := getSchemasPaginated(t, page, 1, http.StatusOK)

				var result []client.IdentitySchemaContainer
				require.NoError(t, json.Unmarshal(body, &result))

				require.Len(t, result, 1)
				assert.Equal(t, actualIDs[page], result[0].Id)
				assertCorrectSchema(t, result[0])
			})
		}
	})

	t.Run("case=read schema", func(t *testing.T) {
		t.Parallel()

		for _, s := range schemas {
			expected, expectedErr := s.getRaw()

			actual, err := reg.SchemaHandler().ReadSchema(t.Context(), urlx.ParseOrPanic(s.uri))
			if expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, expectedErr.Error()) // not using error.is because some of the errors are not accessible
			}

			if expectedErr == nil {
				require.JSONEq(t, string(expected), string(actual))
			}
		}
	})
}
