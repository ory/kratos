package schema_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/ory/jsonschema/v3/fileloader"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
	"github.com/ory/x/urlx"
)

func TestHandler(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	router := x.NewRouterPublic()
	reg.SchemaHandler().RegisterPublicRoutes(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	schemas := schema.Schemas{
		{
			ID:     "default",
			URL:    urlx.ParseOrPanic("file://./stub/identity.schema.json"),
			RawURL: "file://./stub/identity.schema.json",
		},
		{
			ID:     "identity2",
			URL:    urlx.ParseOrPanic("file://./stub/identity-2.schema.json"),
			RawURL: "file://./stub/identity-2.schema.json",
		},
		{
			ID:     "base64",
			URL:    urlx.ParseOrPanic("base64://ewogICIkc2NoZW1hIjogImh0dHA6Ly9qc29uLXNjaGVtYS5vcmcvZHJhZnQtMDcvc2NoZW1hIyIsCiAgInR5cGUiOiAib2JqZWN0IiwKICAicHJvcGVydGllcyI6IHsKICAgICJiYXIiOiB7CiAgICAgICJ0eXBlIjogInN0cmluZyIKICAgIH0KICB9LAogICJyZXF1aXJlZCI6IFsKICAgICJiYXIiCiAgXQp9"),
			RawURL: "base64://ewogICIkc2NoZW1hIjogImh0dHA6Ly9qc29uLXNjaGVtYS5vcmcvZHJhZnQtMDcvc2NoZW1hIyIsCiAgInR5cGUiOiAib2JqZWN0IiwKICAicHJvcGVydGllcyI6IHsKICAgICJiYXIiOiB7CiAgICAgICJ0eXBlIjogInN0cmluZyIKICAgIH0KICB9LAogICJyZXF1aXJlZCI6IFsKICAgICJiYXIiCiAgXQp9",
		},
		{
			ID:     "unreachable",
			URL:    urlx.ParseOrPanic("http://127.0.0.1:12345/unreachable-schema"),
			RawURL: "http://127.0.0.1:12345/unreachable-schema",
		},
		{
			ID:     "no-file",
			URL:    urlx.ParseOrPanic("file://./stub/does-not-exist.schema.json"),
			RawURL: "file://./stub/does-not-exist.schema.json",
		},
		{
			ID:     "directory",
			URL:    urlx.ParseOrPanic("file://./stub"),
			RawURL: "file://./stub",
		},
	}

	getSchemaById := func(id string) *schema.Schema {
		s, err := schemas.GetByID(id)
		require.NoError(t, err)
		return s
	}

	getFromTS := func(url string, expectCode int) []byte {
		res, err := ts.Client().Get(url)
		require.NoError(t, err)
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		require.EqualValues(t, expectCode, res.StatusCode, "%s", body)
		return body

	}

	getFromTSById := func(id string, expectCode int) []byte {
		return getFromTS(fmt.Sprintf("%s/schemas/%s", ts.URL, id), expectCode)
	}

	getFromTSPaginated := func(page, perPage, expectCode int) []byte {
		return getFromTS(fmt.Sprintf("%s/schemas?page=%d&per_page=%d", ts.URL, page, perPage), expectCode)
	}

	getFromFS := func(id string) []byte {
		schema := getSchemaById(id)

		if schema.URL.Scheme == "file" {
			raw, err := os.ReadFile(strings.TrimPrefix(schema.RawURL, "file://"))
			require.NoError(t, err)
			return raw
		} else if schema.URL.Scheme == "base64" {
			data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(schema.RawURL, "base64://"))
			require.NoError(t, err)
			return data
		}
		return nil
	}

	setSchemas := func(newSchemas schema.Schemas) {
		schemas = newSchemas
		var schemasConfig []config.Schema
		for _, s := range schemas {
			if s.ID != config.DefaultIdentityTraitsSchemaID {
				schemasConfig = append(schemasConfig, config.Schema{
					ID:  s.ID,
					URL: s.RawURL,
				})
			}
		}
		conf.MustSet(config.ViperKeyIdentitySchemas, schemasConfig)
	}

	conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, getSchemaById(config.DefaultIdentityTraitsSchemaID).RawURL)
	setSchemas(schemas)

	t.Run("case=get default schema", func(t *testing.T) {
		server := getFromTSById(config.DefaultIdentityTraitsSchemaID, http.StatusOK)
		file := getFromFS(config.DefaultIdentityTraitsSchemaID)
		require.JSONEq(t, string(file), string(server))
	})

	t.Run("case=get other schema", func(t *testing.T) {
		server := getFromTSById("identity2", http.StatusOK)
		file := getFromFS("identity2")
		require.JSONEq(t, string(file), string(server))
	})

	t.Run("case=get base64 schema", func(t *testing.T) {
		server := getFromTSById("base64", http.StatusOK)
		file := getFromFS("base64")
		require.JSONEq(t, string(file), string(server))
	})

	t.Run("case=get unreachable schema", func(t *testing.T) {
		reason := getFromTSById("unreachable", http.StatusInternalServerError)
		require.Contains(t, string(reason), "could not be found or opened")
	})

	t.Run("case=get no-file schema", func(t *testing.T) {
		reason := getFromTSById("no-file", http.StatusInternalServerError)
		require.Contains(t, string(reason), "could not be found or opened")
	})

	t.Run("case=get directory schema", func(t *testing.T) {
		reason := getFromTSById("directory", http.StatusInternalServerError)
		require.Contains(t, string(reason), "could not be found or opened")
	})

	t.Run("case=get not-existing schema", func(t *testing.T) {
		_ = getFromTSById("not-existing", http.StatusNotFound)
	})

	t.Run("case=get all schemas", func(t *testing.T) {
		setSchemas(schema.Schemas{
			{
				ID:     "default",
				URL:    urlx.ParseOrPanic("file://./stub/identity.schema.json"),
				RawURL: "file://./stub/identity.schema.json",
			},
			{
				ID:     "identity2",
				URL:    urlx.ParseOrPanic("file://./stub/identity-2.schema.json"),
				RawURL: "file://./stub/identity-2.schema.json",
			},
		})

		body := getFromTSPaginated(0, 2, http.StatusOK)

		var result schema.IdentitySchemas
		require.NoError(t, json.Unmarshal(body, &result))

		ids_orig := []string{}
		for _, s := range schemas {
			ids_orig = append(ids_orig, s.ID)
		}
		ids_list := []string{}
		for _, s := range result {
			ids_list = append(ids_list, s.ID)
		}
		for _, id := range ids_orig {
			require.Contains(t, ids_list, id)
		}

		for _, s := range schemas {
			for _, r := range result {
				if r.ID == s.ID {
					assert.JSONEq(t, string(getFromFS(s.ID)), string(r.Schema))
				}
			}
		}
	})

	t.Run("case=get paginated schemas", func(t *testing.T) {
		setSchemas(schema.Schemas{
			{
				ID:     "default",
				URL:    urlx.ParseOrPanic("file://./stub/identity.schema.json"),
				RawURL: "file://./stub/identity.schema.json",
			},
			{
				ID:     "identity2",
				URL:    urlx.ParseOrPanic("file://./stub/identity-2.schema.json"),
				RawURL: "file://./stub/identity-2.schema.json",
			},
		})

		body1, body2 := getFromTSPaginated(0, 1, http.StatusOK), getFromTSPaginated(1, 1, http.StatusOK)

		var result1, result2 schema.IdentitySchemas
		require.NoError(t, json.Unmarshal(body1, &result1))
		require.NoError(t, json.Unmarshal(body2, &result2))

		result := append(result1, result2...)

		ids_orig := []string{}
		for _, s := range schemas {
			ids_orig = append(ids_orig, s.ID)
		}
		ids_list := []string{}
		for _, s := range result {
			ids_list = append(ids_list, s.ID)
		}
		for _, id := range ids_orig {
			require.Contains(t, ids_list, id)
		}
	})

	t.Run("case=read schema", func(t *testing.T) {
		setSchemas(schema.Schemas{
			{
				ID:     "default",
				URL:    urlx.ParseOrPanic("file://./stub/identity.schema.json"),
				RawURL: "file://./stub/identity.schema.json",
			},
			{
				ID:     "default",
				URL:    urlx.ParseOrPanic(fmt.Sprintf("%s/schemas/default", ts.URL)),
				RawURL: fmt.Sprintf("%s/schemas/default", ts.URL),
			},
		})

		src, err := schema.ReadSchema(&schemas[0])
		require.NoError(t, err)
		defer src.Close()

		src, err = schema.ReadSchema(&schemas[1])
		require.NoError(t, err)
		defer src.Close()
	})
}
