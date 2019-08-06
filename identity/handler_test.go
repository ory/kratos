package identity_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/viper"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
	"github.com/ory/hive/x"
)

func TestHandler(t *testing.T) {
	_, reg := internal.NewMemoryRegistry(t)
	router := x.NewRouterAdmin()
	reg.IdentityHandler().RegisterAdminRoutes(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	viper.Set(configuration.ViperKeyURLsSelfAdmin, ts.URL)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/identity.schema.json")

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
		require.True(t, parsed.IsArray())
		assert.Len(t, parsed.Array(), 0)
	})

	t.Run("case=should return 404 on a non-existing resource", func(t *testing.T) {
		_ = get(t, "/identities/does-not-exist", http.StatusNotFound)
	})

	t.Run("case=should fail to create an entity because schema url does not exist", func(t *testing.T) {
		var i identity.Identity
		i.TraitsSchemaURL = "file://./stub/does-not-exist.schema.json"
		res := send(t, "POST", "/identities", http.StatusInternalServerError, &i)
		assert.Contains(t, res.Get("error.reason").String(), "no such file or directory")
	})

	t.Run("case=should fail to create an entity because schema is not validating", func(t *testing.T) {
		var i identity.Identity
		i.Traits = json.RawMessage(`{"bar":123}`)
		res := send(t, "POST", "/identities", http.StatusBadRequest, &i)
		assert.Contains(t, res.Get("error.reason").String(), "invalid type")
	})

	t.Run("case=should create an identity without an ID", func(t *testing.T) {
		var i identity.Identity
		i.Traits = json.RawMessage(`{"bar":"baz"}`)
		res := send(t, "POST", "/identities", http.StatusCreated, &i)
		assert.NotEmpty(t, res.Get("id").String(), "%s", res.Raw)
		assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
		assert.Empty(t, res.Get("credentials").String(), "%s", res.Raw)
	})

	t.Run("case=should create an identity with an ID", func(t *testing.T) {
		var i identity.Identity
		i.ID = "exists"
		i.Traits = json.RawMessage(`{"bar":"baz"}`)
		res := send(t, "POST", "/identities", http.StatusCreated, &i)
		assert.EqualValues(t, "exists", res.Get("id").String(), "%s", res.Raw)
		assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
		assert.Empty(t, res.Get("credentials").String(), "%s", res.Raw)
		assert.EqualValues(t, viper.GetString(configuration.ViperKeyDefaultIdentityTraitsSchemaURL), res.Get("traits_schema_url").String(), "%s", res.Raw)
	})

	t.Run("case=should be able to get the identity", func(t *testing.T) {
		res := get(t, "/identities/exists", http.StatusOK)
		assert.EqualValues(t, "exists", res.Get("id").String(), "%s", res.Raw)
		assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
		assert.EqualValues(t, viper.GetString(configuration.ViperKeyDefaultIdentityTraitsSchemaURL), res.Get("traits_schema_url").String(), "%s", res.Raw)
		assert.Empty(t, res.Get("credentials").String(), "%s", res.Raw)
	})

	t.Run("case=should update an identity and persist the changes", func(t *testing.T) {
		var i identity.Identity
		i.Traits = json.RawMessage(`{"bar":"baz","foo":"baz"}`)
		res := send(t, "PUT", "/identities/exists", http.StatusOK, &i)
		assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
		assert.EqualValues(t, "baz", res.Get("traits.foo").String(), "%s", res.Raw)

		res = get(t, "/identities/exists", http.StatusOK)
		assert.EqualValues(t, "exists", res.Get("id").String(), "%s", res.Raw)
		assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
	})

	t.Run("case=should list all identities", func(t *testing.T) {
		res := get(t, "/identities", http.StatusOK)
		assert.Empty(t, res.Get("0.credentials").String(), "%s", res.Raw)
		assert.EqualValues(t, "baz", res.Get("0.traits.bar").String(), "%s", res.Raw)
	})

	t.Run("case=should not be able to update an identity that does not exist yet", func(t *testing.T) {
		var i identity.Identity
		i.ID = uuid.New().String()
		i.Traits = json.RawMessage(`{"bar":"baz"}`)
		res := send(t, "PUT", "/identities/"+i.ID, http.StatusNotFound, &i)
		assert.Contains(t, res.Get("error.reason").String(), "does not exist")
	})

	t.Run("case=should delete a client and no longer be able to retrieve it", func(t *testing.T) {
		remove(t, "/identities/exists", http.StatusNoContent)
		_ = get(t, "/identities/exists", http.StatusNotFound)
	})

	t.Run("case=should return 404 for non-existing clients", func(t *testing.T) {
		remove(t, "/identities/does-not-exist", http.StatusNotFound)
	})
}
