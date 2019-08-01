package x

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession(t *testing.T) {
	const sid = "test_session"

	s := sessions.NewCookieStore([]byte("cyan cat walking over keyboard"))
	cj, err := cookiejar.New(&cookiejar.Options{})
	require.NoError(t, err)
	c := http.Client{Jar: cj}

	router := httprouter.New()
	router.GET("/set", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NoError(t, SessionPersistValues(w, r, s, sid, map[string]interface{}{
			"string-1": "foo",
			"string-2": "bar",
			"string-3": "",
			"int":      1234,
		}))
		w.WriteHeader(http.StatusNoContent)
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	var mr = func(t *testing.T, path string) {
		res, err := c.Get(ts.URL + "/" + path)
		require.NoError(t, err)
		require.EqualValues(t, http.StatusNoContent, res.StatusCode)
		require.NoError(t, res.Body.Close())
	}
	mr(t, "set")

	t.Run("case=GetString", func(t *testing.T) {
		id := "get-string"
		router.GET("/"+id, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			got, err := SessionGetString(r, s, sid, "string-1")
			require.NoError(t, err)
			assert.EqualValues(t, "foo", got)

			got, err = SessionGetString(r, s, sid, "string-2")
			require.NoError(t, err)
			assert.EqualValues(t, "bar", got)

			got, err = SessionGetString(r, s, sid, "string-3")
			require.NoError(t, err)
			assert.EqualValues(t, "", got)

			got, err = SessionGetString(r, s, sid, "int")
			require.Error(t, err)

			got, err = SessionGetString(r, s, sid, "i-dont-exist")
			require.Error(t, err)

			w.WriteHeader(http.StatusNoContent)
		})
		mr(t, id)
	})

	t.Run("case=GetStringOr", func(t *testing.T) {
		id := "get-string-or"
		router.GET("/"+id, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			assert.EqualValues(t, "foo", SessionGetStringOr(r, s, sid, "string-1", "baz"))
			assert.EqualValues(t, "bar", SessionGetStringOr(r, s, sid, "string-2", "baz"))
			assert.EqualValues(t, "", SessionGetStringOr(r, s, sid, "string-3", "baz"))
			assert.EqualValues(t, "baz", SessionGetStringOr(r, s, sid, "int", "baz"))
			assert.EqualValues(t, "baz", SessionGetStringOr(r, s, sid, "i-dont-exist", "baz"))

			w.WriteHeader(http.StatusNoContent)
		})
		mr(t, id)
	})
}
