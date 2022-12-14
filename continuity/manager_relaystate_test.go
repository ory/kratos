// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package continuity_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ory/kratos/driver/config"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/x/ioutilx"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"
	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
)

func TestManagerRelayState(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)

	testhelpers.SetDefaultIdentitySchema(conf, "file://../test/stub/identity/empty.schema.json")
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh")
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
			relayState := r.URL.Query().Get("RelayState")

			r.PostForm = make(url.Values)
			r.PostForm.Set("RelayState", relayState)

			c, err := p.Continue(r.Context(), w, r, ps.ByName("name"), tc.wo...)
			if err != nil {
				writer.WriteError(w, r, err)
				return
			}
			writer.Write(w, r, c)
		})

		router.DELETE("/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			relayState := r.URL.Query().Get("RelayState")

			r.PostForm = make(url.Values)
			r.PostForm.Set("RelayState", relayState)

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

	p := reg.RelayStateContinuityManager()
	cl := newClient()

	t.Run("case=continue cookie persists with same http client", func(t *testing.T) {
		ts := newServer(t, p, new(persisterTestCase))
		name := x.NewUUID().String()
		href := ts.URL + "/" + name

		res, err := cl.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		req := x.NewTestHTTPRequest(t, "POST", href, nil)
		require.Len(t, res.Cookies(), 1)

		res, err = cl.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		body := ioutilx.MustReadAll(res.Body)
		assert.Contains(t, gjson.GetBytes(body, "name").String(), name)

		t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

		require.Len(t, res.Cookies(), 1)
		assert.EqualValues(t, res.Cookies()[0].Name, continuity.CookieName)
	})

	t.Run("case=continue cookie reconstructed and delivered with valid relaystate", func(t *testing.T) {
		ts := newServer(t, p, new(persisterTestCase))
		name := x.NewUUID().String()
		href := ts.URL + "/" + name

		res, err := cl.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		var relayState string

		for _, c := range res.Cookies() {
			relayState = c.Value
		}

		req := x.NewTestHTTPRequest(t, "POST", href+"?RelayState="+url.QueryEscape(relayState), nil)
		require.Len(t, res.Cookies(), 1)

		res, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		body := ioutilx.MustReadAll(res.Body)
		assert.Contains(t, gjson.GetBytes(body, "name").String(), name)

		t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

		require.Len(t, res.Cookies(), 1)
		assert.EqualValues(t, res.Cookies()[0].Name, continuity.CookieName)
	})

	t.Run("case=continue cookie not delivered with invalid relaystate", func(t *testing.T) {
		ts := newServer(t, p, new(persisterTestCase))
		name := x.NewUUID().String()
		href := ts.URL + "/" + name

		res, err := cl.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		var relayState string

		for _, c := range res.Cookies() {
			relayState = c.Value
			relayState = strings.Replace(relayState, "a", "b", 1)
		}
		require.Len(t, res.Cookies(), 1)

		req := x.NewTestHTTPRequest(t, "POST", href+"?RelayState="+url.QueryEscape(relayState), nil)

		res, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		body := ioutilx.MustReadAll(res.Body)
		assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), continuity.ErrNotResumable.ReasonField)

		t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

		require.Len(t, res.Cookies(), 0, "the cookie couldn't be reconstructed without a valid relaystate")
	})

	t.Run("case=continue cookie not delivered without relaystate", func(t *testing.T) {
		ts := newServer(t, p, new(persisterTestCase))
		name := x.NewUUID().String()
		href := ts.URL + "/" + name

		res, err := cl.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusNoContent, res.StatusCode)
		require.Len(t, res.Cookies(), 1)

		req := x.NewTestHTTPRequest(t, "POST", href, nil)

		res, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		body := ioutilx.MustReadAll(res.Body)
		assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), continuity.ErrNotResumable.ReasonField)

		t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

		require.Len(t, res.Cookies(), 0, "the cookie couldn't be reconstructed without a valid relaystate")
	})

	t.Run("case=pause, abort, and continue session with failure", func(t *testing.T) {
		ts := newServer(t, p, new(persisterTestCase))
		name := x.NewUUID().String()
		href := ts.URL + "/" + name

		res, err := cl.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		req := x.NewTestHTTPRequest(t, "DELETE", href, nil)

		res, err = cl.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, res.Body.Close()) })
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		req = x.NewTestHTTPRequest(t, "POST", href, nil)

		res, err = cl.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		body := ioutilx.MustReadAll(res.Body)
		assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), continuity.ErrNotResumable.ReasonField)

		t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

		require.Len(t, res.Cookies(), 0, "the cookie couldn't be reconstructed without a valid relaystate")
	})
}
