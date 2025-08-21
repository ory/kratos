// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"net/http"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/urfave/negroni"

	"github.com/ory/x/httprouterx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouterAdmin(t *testing.T) {
	require.NotEmpty(t, NewTestRouterAdmin(t))
	require.NotEmpty(t, NewTestRouterPublic(t))
}

func TestCacheHandling(t *testing.T) {
	router := NewTestRouterPublic(t)
	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	router.HandleFunc("GET /foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.HandleFunc("DELETE /foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.HandleFunc("POST /foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.HandleFunc("PUT /foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.HandleFunc("PATCH /foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	for _, method := range []string{"GET", "DELETE", "POST", "PUT", "PATCH"} {
		req, _ := http.NewRequest(method, ts.URL+"/foo", nil)
		res, err := ts.Client().Do(req)
		require.NoError(t, err)
		assert.EqualValues(t, "private, no-cache, no-store, must-revalidate", res.Header.Get("Cache-Control"))
	}
}

func TestAdminPrefix(t *testing.T) {
	n := negroni.New()
	n.UseFunc(httprouterx.TrimTrailingSlashNegroni)
	n.UseFunc(httprouterx.NoCacheNegroni)
	n.UseFunc(httprouterx.AddAdminPrefixIfNotPresentNegroni)

	router := NewTestRouterAdmin(t)
	n.UseHandler(router)

	ts := httptest.NewServer(n)
	t.Cleanup(ts.Close)

	router.HandleFunc("GET /admin/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.HandleFunc("DELETE /admin/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.HandleFunc("POST /admin/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.HandleFunc("PUT /admin/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.HandleFunc("PATCH /admin/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	for _, method := range []string{"GET", "DELETE", "POST", "PUT", "PATCH"} {
		{
			req, _ := http.NewRequest(method, ts.URL+"/foo", nil)
			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusNoContent, res.StatusCode)
		}
		{
			req, _ := http.NewRequest(method, ts.URL+"/admin/foo", nil)
			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusNoContent, res.StatusCode)
		}
	}
}
