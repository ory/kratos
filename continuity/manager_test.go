package continuity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ory/x/ioutilx"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"
	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
)

type persisterTestCase struct {
	ro          []continuity.ManagerOption
	wo          []continuity.ManagerOption
	expected    interface{}
	expectedErr error
}

type persisterTestPayload struct {
	Foo string `json:"foo"`
}

func TestManager(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://../test/stub/identity/empty.schema.json")
	conf.MustSet(config.ViperKeyPublicBaseURL, "https://www.ory.sh")
	i := identity.NewIdentity("")
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

	var newServer = func(t *testing.T, p continuity.Manager, tc *persisterTestCase) *httptest.Server {
		writer := herodot.NewJSONWriter(logrusx.New("", ""))
		router := httprouter.New()
		router.PUT("/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			if err := p.Pause(r.Context(), w, r, ps.ByName("name"), tc.ro...); err != nil {
				writer.WriteError(w, r, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		router.POST("/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			if err := p.Pause(r.Context(), w, r, ps.ByName("name"), tc.ro...); err != nil {
				writer.WriteError(w, r, err)
				return
			}

			c, err := p.Continue(r.Context(), w, r, ps.ByName("name"), tc.wo...)
			if err != nil {
				writer.WriteError(w, r, err)
				return
			}
			writer.Write(w, r, c)
		})

		router.GET("/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			c, err := p.Continue(r.Context(), w, r, ps.ByName("name"), tc.ro...)
			if err != nil {
				writer.WriteError(w, r, err)
				return
			}
			writer.Write(w, r, c)
		})

		router.DELETE("/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			err := p.Abort(r.Context(), w, r, ps.ByName("name"))
			if err != nil {
				writer.WriteError(w, r, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		ts := httptest.NewServer(router)
		t.Cleanup(func() {
			ts.Close()
		})
		return ts
	}

	var newClient = func() *http.Client {
		return &http.Client{Jar: x.EasyCookieJar(t, nil)}
	}

	p := reg.ContinuityManager()
	cl := newClient()

	t.Run("case=continue cookie resets when signature is invalid", func(t *testing.T) {
		ts := newServer(t, p, new(persisterTestCase))
		href := ts.URL + "/" + x.NewUUID().String()

		res, err := cl.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		req := x.NewTestHTTPRequest(t, "GET", href, nil)
		require.Len(t, res.Cookies(), 1)
		for _, c := range res.Cookies() {
			// Change something in the string
			c.Value = strings.Replace(c.Value, "a", "b", 1)
			req.AddCookie(c)
		}
		res, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

		require.Equal(t, http.StatusBadRequest, res.StatusCode)
		body := ioutilx.MustReadAll(res.Body)
		assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), continuity.ErrNotResumable.ReasonField)

		require.Len(t, res.Cookies(), 1, "continuing the flow with a broken cookie should instruct the browser to forget it")
		assert.EqualValues(t, res.Cookies()[0].Name, continuity.CookieName)
	})

	t.Run("case=can deal with duplicate cookies", func(t *testing.T) {
		tc := &persisterTestCase{expected: &persisterTestPayload{"bar"}}
		ts := newServer(t, p, tc)
		href := ts.URL + "/" + x.NewUUID().String()

		res, err := http.DefaultClient.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		// We change the key to another one
		href = ts.URL + "/" + x.NewUUID().String()
		req := x.NewTestHTTPRequest(t, "GET", href, nil)
		require.Len(t, res.Cookies(), 1)
		for _, c := range res.Cookies() {
			req.AddCookie(c)
		}

		tc.ro = []continuity.ManagerOption{continuity.WithPayload(&persisterTestPayload{"bar"})}
		res, err = http.DefaultClient.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		require.Len(t, res.Cookies(), 1)
		for _, c := range res.Cookies() {
			req.AddCookie(c)
		}

		res, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

		require.Len(t, res.Cookies(), 1, "continuing the flow with a broken cookie should instruct the browser to forget it")
		assert.EqualValues(t, res.Cookies()[0].Name, continuity.CookieName)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(tc.expected))
		body := ioutilx.MustReadAll(res.Body)
		assert.JSONEq(t, b.String(), gjson.GetBytes(body, "payload").Raw, "%s", body)
		assert.Contains(t, href, gjson.GetBytes(body, "name").String(), "%s", body)
	})

	for k, tc := range []persisterTestCase{
		{},
		{
			ro:       []continuity.ManagerOption{continuity.WithPayload(&persisterTestPayload{"bar"})},
			wo:       []continuity.ManagerOption{continuity.WithPayload(&persisterTestPayload{})},
			expected: &persisterTestPayload{"bar"},
		},
		{
			ro: []continuity.ManagerOption{continuity.WithIdentity(i)},
			wo: []continuity.ManagerOption{continuity.WithIdentity(i)},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			cl := newClient()
			ts := newServer(t, p, &tc)
			var genid = func() string {
				return ts.URL + "/" + x.NewUUID().String()
			}

			t.Run("case=resume non-existing session", func(t *testing.T) {
				href := genid()
				res, err := cl.Do(x.NewTestHTTPRequest(t, "GET", href, nil))
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

				body := ioutilx.MustReadAll(res.Body)
				require.Equal(t, http.StatusBadRequest, res.StatusCode)
				assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), continuity.ErrNotResumable.ReasonField)
			})

			t.Run("case=pause and resume session", func(t *testing.T) {
				href := genid()
				res, err := cl.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
				require.NoError(t, err)
				require.NoError(t, res.Body.Close())
				require.Equal(t, http.StatusNoContent, res.StatusCode)

				res, err = cl.Do(x.NewTestHTTPRequest(t, "GET", href, nil))
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

				body := ioutilx.MustReadAll(res.Body)
				if tc.expectedErr != nil {
					require.Equal(t, http.StatusGone, res.StatusCode, "%s", body)
					return
				}

				require.Equal(t, http.StatusOK, res.StatusCode, "%s", body)

				var b bytes.Buffer
				require.NoError(t, json.NewEncoder(&b).Encode(tc.expected))
				assert.JSONEq(t, b.String(), gjson.GetBytes(body, "payload").Raw, "%s", body)
				assert.Contains(t, href, gjson.GetBytes(body, "name").String(), "%s", body)
			})

			t.Run("case=pause and retry session", func(t *testing.T) {
				href := genid()
				res, err := cl.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
				require.NoError(t, err)
				require.NoError(t, res.Body.Close())
				require.Equal(t, http.StatusNoContent, res.StatusCode)

				res, err = cl.Do(x.NewTestHTTPRequest(t, "GET", href, nil))
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

				res, err = cl.Do(x.NewTestHTTPRequest(t, "GET", href, nil))
				require.NoError(t, err)
				require.Equal(t, http.StatusBadRequest, res.StatusCode)
				body := ioutilx.MustReadAll(res.Body)
				t.Cleanup(func() { require.NoError(t, res.Body.Close()) })
				assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), continuity.ErrNotResumable.ReasonField)
			})

			t.Run("case=pause and resume session in the same request", func(t *testing.T) {
				href := genid()
				res, err := cl.Do(x.NewTestHTTPRequest(t, "POST", href, nil))
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, res.StatusCode)
				t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

				var b bytes.Buffer
				require.NoError(t, json.NewEncoder(&b).Encode(tc.expected))

				body := ioutilx.MustReadAll(res.Body)
				assert.JSONEq(t, b.String(), gjson.GetBytes(body, "payload").Raw, "%s", body)
				assert.Contains(t, href, gjson.GetBytes(body, "name").String(), "%s", body)
			})

			t.Run("case=pause, abort, and continue session with failure", func(t *testing.T) {
				href := genid()
				res, err := cl.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
				require.NoError(t, err)
				require.NoError(t, res.Body.Close())
				require.Equal(t, http.StatusNoContent, res.StatusCode)

				res, err = cl.Do(x.NewTestHTTPRequest(t, "DELETE", href, nil))
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, res.Body.Close()) })
				require.Equal(t, http.StatusNoContent, res.StatusCode)

				res, err = cl.Do(x.NewTestHTTPRequest(t, "GET", href, nil))
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

				require.Equal(t, http.StatusBadRequest, res.StatusCode)
				body := ioutilx.MustReadAll(res.Body)
				assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), continuity.ErrNotResumable.ReasonField)
			})
		})
	}
}
