package identity_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
)

func TestHandler(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
	router := x.NewRouterAdmin()
	reg.IdentityHandler().RegisterAdminRoutes(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	mockServerURL := urlx.ParseOrPanic("http://example.com")
	defaultSchemaInternalURL := urlx.ParseOrPanic("file://./stub/identity.schema.json")
	defaultSchema := schema.Schema{
		ID:  "default",
		URL: defaultSchemaInternalURL,
	}
	defaultSchemaExternalURL := defaultSchema.SchemaURL(mockServerURL).String()

	viper.Set(configuration.ViperKeyAdminBaseURL, ts.URL)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, defaultSchemaInternalURL.String())
	viper.Set(configuration.ViperKeyPublicBaseURL, mockServerURL.String())

	var get = func(t *testing.T, href string, expectCode int) gjson.Result {
		res, err := ts.Client().Get(ts.URL + href)
		require.NoError(t, err)
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		require.EqualValues(t, expectCode, res.StatusCode, "%s", body)
		return gjson.ParseBytes(body)
	}

	var remove = func(t *testing.T, href string, expectCode int) {
		req, err := http.NewRequest("DELETE", ts.URL+href, nil)
		require.NoError(t, err)

		res, err := ts.Client().Do(req)
		require.NoError(t, err)

		require.EqualValues(t, expectCode, res.StatusCode)
	}

	var send = func(t *testing.T, method, href string, expectCode int, send interface{}) gjson.Result {
		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(send))
		req, err := http.NewRequest(method, ts.URL+href, &b)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		res, err := ts.Client().Do(req)
		require.NoError(t, err)
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		require.EqualValues(t, expectCode, res.StatusCode, "%s", body)
		return gjson.ParseBytes(body)
	}

	t.Run("case=should return an empty list", func(t *testing.T) {
		parsed := get(t, "/identities", http.StatusOK)
		require.True(t, parsed.IsArray(), "%s", parsed.Raw)
		assert.Len(t, parsed.Array(), 0)
	})

	t.Run("case=should return 404 on a non-existing resource", func(t *testing.T) {
		_ = get(t, "/identities/does-not-exist", http.StatusNotFound)
	})

	t.Run("case=should fail to create an identity because schema id does not exist", func(t *testing.T) {
		var i identity.Identity
		i.TraitsSchemaID = "does-not-exist"
		res := send(t, "POST", "/identities", http.StatusBadRequest, &i)
		assert.Contains(t, res.Get("error.reason").String(), "does-not-exist", "%s", res)
	})

	t.Run("case=should fail to create an entity because schema is not validating", func(t *testing.T) {
		var i identity.Identity
		i.Traits = identity.Traits(`{"bar":123}`)
		res := send(t, "POST", "/identities", http.StatusBadRequest, &i)
		assert.Contains(t, res.Get("error.reason").String(), "I[#/traits/bar] S[#/properties/bar/type] expected string, but got number")
	})

	t.Run("case=should fail to create an entity with traits_schema_url set", func(t *testing.T) {
		var i identity.Identity
		i.TraitsSchemaURL = "http://example.com"
		res := send(t, "POST", "/identities", http.StatusBadRequest, &i)
		assert.Contains(t, res.Get("error.reason").String(), "set a traits schema")
	})

	t.Run("case=should create an identity without an ID", func(t *testing.T) {
		var i identity.Identity
		i.Traits = identity.Traits(`{"bar":"baz"}`)
		res := send(t, "POST", "/identities", http.StatusCreated, &i)
		assert.NotEmpty(t, res.Get("id").String(), "%s", res.Raw)
		assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
		assert.Empty(t, res.Get("credentials").String(), "%s", res.Raw)
	})

	var i identity.Identity
	t.Run("case=should create an identity with an ID which is ignored", func(t *testing.T) {
		i.ID = x.NewUUID()
		i.Traits = identity.Traits(`{"bar":"baz"}`)
		res := send(t, "POST", "/identities", http.StatusCreated, &i)
		assert.NotEqual(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)

		i.ID = x.ParseUUID(res.Get("id").String())
		assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
		assert.Empty(t, res.Get("credentials").String(), "%s", res.Raw)
		assert.EqualValues(t, defaultSchemaExternalURL, res.Get("traits_schema_url").String(), "%s", res.Raw)
		assert.EqualValues(t, configuration.DefaultIdentityTraitsSchemaID, res.Get("traits_schema_id").String(), "%s", res.Raw)
	})

	t.Run("case=should be able to get the identity", func(t *testing.T) {
		res := get(t, "/identities/"+i.ID.String(), http.StatusOK)
		assert.EqualValues(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)
		assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
		assert.EqualValues(t, defaultSchemaExternalURL, res.Get("traits_schema_url").String(), "%s", res.Raw)
		assert.EqualValues(t, configuration.DefaultIdentityTraitsSchemaID, res.Get("traits_schema_id").String(), "%s", res.Raw)
		assert.Empty(t, res.Get("credentials").String(), "%s", res.Raw)
	})

	t.Run("case=should update an identity and persist the changes", func(t *testing.T) {
		i.Traits = identity.Traits(`{"bar":"baz","foo":"baz"}`)
		res := send(t, "PUT", "/identities/"+i.ID.String(), http.StatusOK, &i)
		assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
		assert.EqualValues(t, "baz", res.Get("traits.foo").String(), "%s", res.Raw)

		res = get(t, "/identities/"+i.ID.String(), http.StatusOK)
		assert.EqualValues(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)
		assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
	})

	t.Run("case=should list all identities", func(t *testing.T) {
		res := get(t, "/identities", http.StatusOK)
		assert.Empty(t, res.Get("0.credentials").String(), "%s", res.Raw)
		assert.EqualValues(t, "baz", res.Get("0.traits.bar").String(), "%s", res.Raw)
	})

	t.Run("case=should not be able to update an identity that does not exist yet", func(t *testing.T) {
		var i identity.Identity
		i.ID = x.NewUUID()
		i.Traits = identity.Traits(`{"bar":"baz"}`)
		res := send(t, "PUT", "/identities/"+i.ID.String(), http.StatusNotFound, &i)
		assert.Contains(t, res.Get("error.message").String(), "Unable to locate the resource", "%s", res.Raw)
	})

	t.Run("case=should delete a client and no longer be able to retrieve it", func(t *testing.T) {
		remove(t, "/identities/"+i.ID.String(), http.StatusNoContent)
		_ = get(t, "/identities/"+i.ID.String(), http.StatusNotFound)
	})

	t.Run("case=should return 404 for non-existing identities", func(t *testing.T) {
		remove(t, "/identities/"+x.NewUUID().String(), http.StatusNotFound)
	})
}
