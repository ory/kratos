package x

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession(t *testing.T) {
	const sid = "test_session"

	s := sessions.NewCookieStore([]byte("cyan cat walking over keyboard"))
	s.Options.MaxAge = 78652871
	cj, err := cookiejar.New(&cookiejar.Options{})
	require.NoError(t, err)
	c := http.Client{Jar: cj}

	isExpiryCorrect := func(t *testing.T, r *http.Request) {
		cookie, _ := s.Get(r, sid)
		assert.EqualValues(t, 78652871, cookie.Options.MaxAge, "we ensure the options are always copied correctly.")
	}

	router := httprouter.New()
	router.GET("/set", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NoError(t, SessionPersistValues(w, r, s, sid, map[string]interface{}{
			"string-1": "foo",
			"string-2": "bar",
			"string-3": "",
			"int":      1234,
		}))
		isExpiryCorrect(t, r)
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

			_, err = SessionGetString(r, s, sid, "int")
			require.Error(t, err)

			_, err = SessionGetString(r, s, sid, "i-dont-exist")
			require.Error(t, err)

			w.WriteHeader(http.StatusNoContent)
		})
	})

	t.Run("case=GetStringMultipleCookies", func(t *testing.T) {
		id := "get-string-multiple"

		router.GET("/set/"+id, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			require.NoError(t, SessionPersistValues(w, r, s, sid, map[string]interface{}{
				"multiple-string-1": "foo",
			}))
			require.NoError(t, SessionPersistValues(w, r, s, sid, map[string]interface{}{
				"multiple-string-2": "bar",
			}))
			isExpiryCorrect(t, r)
			w.WriteHeader(http.StatusNoContent)
		})

		router.GET("/get/"+id, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			got, err := SessionGetString(r, s, sid, "multiple-string-1")
			require.NoError(t, err)
			assert.EqualValues(t, "foo", got)

			got, err = SessionGetString(r, s, sid, "multiple-string-2")
			require.NoError(t, err)
			assert.EqualValues(t, "bar", got)

			w.WriteHeader(http.StatusNoContent)
		})

		res, err := http.DefaultClient.Get(ts.URL + "/set/" + id)
		require.NoError(t, err)
		require.EqualValues(t, http.StatusNoContent, res.StatusCode)
		require.NoError(t, res.Body.Close())

		req, _ := http.NewRequest("GET", ts.URL+"/get/"+id, nil)
		for _, c := range res.Cookies() {
			req.AddCookie(c)
		}

		res, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.EqualValues(t, http.StatusNoContent, res.StatusCode)
		require.NoError(t, res.Body.Close())
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

	mrCookie := func(t *testing.T, loc string, cks []*http.Cookie, f func(c []*http.Cookie, req *http.Request)) *http.Response {
		req, err := http.NewRequest("GET", loc, nil)
		require.NoError(t, err)

		// Mess with the signature
		if f != nil {
			f(cks, req)
		}

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return res
	}

	mrUnset := func(t *testing.T, ts *httptest.Server, id string, f func(c []*http.Cookie, req *http.Request)) *http.Response {
		res, err := http.Get(ts.URL + "/set")
		require.NoError(t, err)

		res = mrCookie(t, ts.URL+"/"+id+"/unset", res.Cookies(), f)
		require.EqualValues(t, http.StatusNoContent, res.StatusCode)
		require.NoError(t, res.Body.Close())
		return res
	}

	signatureScrambler := func(c []*http.Cookie, req *http.Request) {
		for _, c := range c {
			c.Value = strings.Replace(c.Value, "a", "b", -1)
			req.AddCookie(c)
		}
	}

	cookieCopier := func(c []*http.Cookie, req *http.Request) {
		for _, c := range c {
			req.AddCookie(c)
		}
	}

	t.Run("case=SessionSet", func(t *testing.T) {
		res, err := http.Get(ts.URL + "/set")
		require.NoError(t, err)

		res = mrCookie(t, ts.URL+"/set", res.Cookies(), signatureScrambler)
		require.EqualValues(t, http.StatusNoContent, res.StatusCode)
		require.NoError(t, res.Body.Close())
	})

	t.Run("case=SessionUnset", func(t *testing.T) {
		t.Run("prevent https://github.com/gorilla/sessions/pull/251", func(t *testing.T) {
			w := httptest.NewRecorder()
			require.NoError(t, SessionUnset(w, new(http.Request), s, "idonotexist"))
			require.NoError(t, SessionUnset(w, new(http.Request), s, "idonotexist"))
			require.Len(t, w.Result().Cookies(), 0)
		})

		id := "session-unset"
		router.GET("/"+id+"/unset", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			require.NoError(t, SessionUnset(w, r, s, sid))
			w.WriteHeader(http.StatusNoContent)
			cookie, _ := s.Get(r, sid)
			assert.EqualValues(t, -1, cookie.Options.MaxAge, "we ensure the options are always copied correctly.")
		})

		router.GET("/"+id+"/get", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			require.Empty(t, SessionGetStringOr(r, s, sid, "string-1", ""))
			require.Empty(t, SessionGetStringOr(r, s, sid, "string-2", ""))
			require.Empty(t, SessionGetStringOr(r, s, sid, "string-3", ""))
			require.Empty(t, SessionGetStringOr(r, s, sid, "int", ""))
			require.Empty(t, SessionGetStringOr(r, s, sid, "i-dont-exist", ""))
			w.WriteHeader(http.StatusNoContent)
		})

		t.Run("with invalid cookie signature", func(t *testing.T) {
			res := mrUnset(t, ts, id, signatureScrambler)
			assert.Len(t, res.Cookies(), 1)
			mrCookie(t, ts.URL+"/"+id+"/get", res.Cookies(), cookieCopier)
			require.EqualValues(t, http.StatusNoContent, res.StatusCode)
			require.NoError(t, res.Body.Close())
		})

		t.Run("with valid cookie signature", func(t *testing.T) {
			res := mrUnset(t, ts, id, cookieCopier)
			assert.Len(t, res.Cookies(), 1)
			mrCookie(t, ts.URL+"/"+id+"/get", res.Cookies(), cookieCopier)
			require.EqualValues(t, http.StatusNoContent, res.StatusCode)
			require.NoError(t, res.Body.Close())
		})
	})

	t.Run("case=SessionUnsetKey", func(t *testing.T) {
		t.Run("prevent https://github.com/gorilla/sessions/pull/251", func(t *testing.T) {
			w := httptest.NewRecorder()
			require.NoError(t, SessionUnsetKey(w, new(http.Request), s, "idonotexist", ""))
			require.NoError(t, SessionUnsetKey(w, new(http.Request), s, "idonotexist", ""))
			require.Len(t, w.Result().Cookies(), 0)
		})

		id := "session-unset-key"
		router.GET("/"+id+"/unset", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			require.NoError(t, SessionUnsetKey(w, r, s, sid, "string-1"))
			w.WriteHeader(http.StatusNoContent)
			isExpiryCorrect(t, r)
		})

		router.GET("/"+id+"/expect-unset", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			require.Empty(t, SessionGetStringOr(r, s, sid, "string-1", ""))
			require.Empty(t, SessionGetStringOr(r, s, sid, "string-2", ""))
			require.Empty(t, SessionGetStringOr(r, s, sid, "string-3", ""))
			require.Empty(t, SessionGetStringOr(r, s, sid, "int", ""))
			require.Empty(t, SessionGetStringOr(r, s, sid, "i-dont-exist", ""))
			w.WriteHeader(http.StatusNoContent)
		})

		router.GET("/"+id+"/expect-one", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			require.Empty(t, SessionGetStringOr(r, s, sid, "string-1", ""))
			assert.EqualValues(t, "bar", SessionGetStringOr(r, s, sid, "string-2", "baz"))
			assert.EqualValues(t, "", SessionGetStringOr(r, s, sid, "string-3", "baz"))
			assert.EqualValues(t, "baz", SessionGetStringOr(r, s, sid, "int", "baz"))
			assert.EqualValues(t, "baz", SessionGetStringOr(r, s, sid, "i-dont-exist", "baz"))
			w.WriteHeader(http.StatusNoContent)
		})

		t.Run("with invalid cookie signature", func(t *testing.T) {
			res := mrUnset(t, ts, id, signatureScrambler)
			assert.Len(t, res.Cookies(), 1)
			mrCookie(t, ts.URL+"/"+id+"/expect-unset", res.Cookies(), cookieCopier)
			require.EqualValues(t, http.StatusNoContent, res.StatusCode)
			require.NoError(t, res.Body.Close())
		})

		t.Run("with valid cookie signature", func(t *testing.T) {
			res := mrUnset(t, ts, id, cookieCopier)
			assert.Len(t, res.Cookies(), 1)
			mrCookie(t, ts.URL+"/"+id+"/expect-one", res.Cookies(), cookieCopier)
			require.EqualValues(t, http.StatusNoContent, res.StatusCode)
			require.NoError(t, res.Body.Close())
		})
	})
}
