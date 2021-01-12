package schema_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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

	getFromTS := func(id string, expectCode int) string {
		res, err := ts.Client().Get(fmt.Sprintf("%s/schemas/%s", ts.URL, id))
		require.NoError(t, err)
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		require.EqualValues(t, expectCode, res.StatusCode, "%s", body)
		return string(body)
	}

	getFromFS := func(id string) string {
		f, err := os.Open(strings.TrimPrefix(getSchemaById(id).RawURL, "file://"))
		require.NoError(t, err)
		raw, err := ioutil.ReadAll(f)
		require.NoError(t, err)
		require.NoError(t, f.Close())
		return string(raw)
	}

	var schemasConfig []config.Schema
	for _, s := range schemas {
		if s.ID != config.DefaultIdentityTraitsSchemaID {
			schemasConfig = append(schemasConfig, config.Schema{
				ID:  s.ID,
				URL: s.RawURL,
			})
		}
	}

	conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)
	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, getSchemaById(config.DefaultIdentityTraitsSchemaID).RawURL)
	conf.MustSet(config.ViperKeyIdentitySchemas, schemasConfig)

	t.Run("case=get default schema", func(t *testing.T) {
		server := getFromTS(config.DefaultIdentityTraitsSchemaID, http.StatusOK)
		file := getFromFS(config.DefaultIdentityTraitsSchemaID)
		require.Equal(t, file, server)
	})

	t.Run("case=get other schema", func(t *testing.T) {
		server := getFromTS("identity2", http.StatusOK)
		file := getFromFS("identity2")
		require.Equal(t, file, server)
	})

	t.Run("case=get unreachable schema", func(t *testing.T) {
		reason := getFromTS("unreachable", http.StatusInternalServerError)
		require.Contains(t, reason, "could not be found or opened")
	})

	t.Run("case=get no-file schema", func(t *testing.T) {
		reason := getFromTS("no-file", http.StatusInternalServerError)
		require.Contains(t, reason, "could not be found or opened")
	})

	t.Run("case=get directory schema", func(t *testing.T) {
		reason := getFromTS("directory", http.StatusInternalServerError)
		require.Contains(t, reason, "could not be found or opened")
	})

	t.Run("case=get not-existing schema", func(t *testing.T) {
		_ = getFromTS("not-existing", http.StatusNotFound)
	})
}
