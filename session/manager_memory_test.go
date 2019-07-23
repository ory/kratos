package session_test

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/hive/internal"
	. "github.com/ory/hive/session"
)

func init() {
	internal.RegisterFakes()
}

func TestSessionManager(t *testing.T) {
	conf, reg := internal.NewMemoryRegistry(t)
	sm := NewManagerMemory(conf, reg)

	_, err := sm.Get("does-not-exist")
	require.Error(t, err)

	var gave Session
	require.NoError(t, faker.FakeData(&gave))

	require.NoError(t, sm.Create(&gave))

	got, err := sm.Get(gave.SID)
	require.NoError(t, err)
	assert.EqualValues(t, &gave, got)

	require.NoError(t, sm.Delete(gave.SID))
	_, err = sm.Get(gave.SID)
	require.Error(t, err)
}

func TestSessionManagerHTTP(t *testing.T) {
	conf, reg := internal.NewMemoryRegistry(t)
	sm := NewManagerMemory(conf, reg)
	h := herodot.NewJSONWriter(nil)

	var s Session
	require.NoError(t, faker.FakeData(&s))

	router := httprouter.New()
	router.GET("/set", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NoError(t, sm.Create(&s))
		require.NoError(t, sm.SaveToRequest(&s, w, r))
	})

	router.GET("/set-direct", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, err := sm.CreateToRequest(s.Identity, w, r)
		require.NoError(t, err)
	})

	router.GET("/clear", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NoError(t, sm.PurgeFromRequest(w, r))
	})

	router.GET("/get", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s, err := sm.FetchFromRequest(r)
		if errors.Cause(err) == ErrNoActiveSessionFound {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(t, err)
			return
		}

		h.Write(w, r, s)
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("case=exists", func(t *testing.T) {
		c := &http.Client{}
		c.Jar, _ = cookiejar.New(nil)

		res, err := c.Get(ts.URL + "/set")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		res, err = c.Get(ts.URL + "/get")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		var got Session
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
	})

	t.Run("case=exists-direct", func(t *testing.T) {
		c := &http.Client{}
		c.Jar, _ = cookiejar.New(nil)

		res, err := c.Get(ts.URL + "/set-direct")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		res, err = c.Get(ts.URL + "/get")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		var got Session
		require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
	})

	t.Run("case=exists_clears", func(t *testing.T) {
		c := &http.Client{}
		c.Jar, _ = cookiejar.New(nil)

		res, err := c.Get(ts.URL + "/set")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		res, err = c.Get(ts.URL + "/clear")
		require.NoError(t, err)
		defer res.Body.Close()
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		res, err = c.Get(ts.URL + "/get")
		require.NoError(t, err)
		defer res.Body.Close()
		assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("case=not-exist", func(t *testing.T) {
		c := &http.Client{}
		c.Jar, _ = cookiejar.New(nil)

		res, err := c.Get(ts.URL + "/get")
		require.NoError(t, err)
		defer res.Body.Close()
		assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
	})
}
