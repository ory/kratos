package errorx_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/nosurf"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/errorsx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
)

func TestHandler(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)
	h := errorx.NewHandler(reg)

	t.Run("case=public authorization", func(t *testing.T) {
		router := x.NewRouterPublic()
		ns := x.NewTestCSRFHandler(router, reg)

		h.RegisterPublicRoutes(router)
		router.GET("/regen", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			ns.RegenerateToken(w, r)
			w.WriteHeader(http.StatusNoContent)
		})
		router.GET("/set-error", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			id, err := reg.SelfServiceErrorPersister().Add(context.Background(), nosurf.Token(r), herodot.ErrNotFound.WithReason("foobar"))
			require.NoError(t, err)
			_, _ = w.Write([]byte(id.String()))
		})

		ts := httptest.NewServer(ns)
		defer ts.Close()

		getBody := func(t *testing.T, hc *http.Client, path string, expectedCode int) []byte {
			res, err := hc.Get(ts.URL + path)
			require.NoError(t, err)
			defer res.Body.Close()
			require.EqualValues(t, expectedCode, res.StatusCode)
			body, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			return body
		}
		expectedError := x.MustEncodeJSON(t, []error{herodot.ErrNotFound.WithReason("foobar")})

		t.Run("call with valid csrf cookie", func(t *testing.T) {
			jar, _ := cookiejar.New(nil)
			hc := &http.Client{Jar: jar}
			id := getBody(t, hc, "/set-error", http.StatusOK)
			actual := getBody(t, hc, errorx.ErrorsPath+"?error="+string(id), http.StatusOK)
			assert.JSONEq(t, expectedError, gjson.GetBytes(actual, "errors").Raw, "%s", actual)

			// We expect a forbid error if the error is not found, regardless of CSRF
			_ = getBody(t, hc, errorx.ErrorsPath+"?error=does-not-exist", http.StatusForbidden)
		})

		t.Run("call without any cookies", func(t *testing.T) {
			hc := &http.Client{}
			id := getBody(t, hc, "/set-error", http.StatusOK)
			_ = getBody(t, hc, errorx.ErrorsPath+"?error="+string(id), http.StatusForbidden)
		})

		t.Run("call with different csrf cookie", func(t *testing.T) {
			jar, _ := cookiejar.New(nil)
			hc := &http.Client{Jar: jar}
			id := getBody(t, hc, "/set-error", http.StatusOK)
			_ = getBody(t, hc, "/regen", http.StatusNoContent)
			_ = getBody(t, hc, errorx.ErrorsPath+"?error="+string(id), http.StatusForbidden)
		})
	})

	t.Run("case=errors types", func(t *testing.T) {
		router := x.NewRouterAdmin()
		h.RegisterAdminRoutes(router)
		ts := httptest.NewServer(router)
		defer ts.Close()

		for k, tc := range []struct {
			gave []error
		}{
			{
				gave: []error{
					herodot.ErrNotFound.WithReason("foobar"),
				},
			},
			{
				gave: []error{
					herodot.ErrNotFound.WithReason("foobar"),
					herodot.ErrNotFound.WithReason("foobar"),
				},
			},
			{
				gave: []error{
					herodot.ErrNotFound.WithReason("foobar"),
				},
			},
			{
				gave: []error{
					errors.WithStack(herodot.ErrNotFound.WithReason("foobar")),
				},
			},
			{
				gave: []error{
					errors.WithStack(herodot.ErrNotFound.WithReason("foobar").WithTrace(errors.New("asdf"))),
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				csrf := x.NewUUID()
				id, err := reg.SelfServiceErrorPersister().Add(context.Background(), csrf.String(), tc.gave...)
				require.NoError(t, err)

				res, err := ts.Client().Get(ts.URL + errorx.ErrorsPath + "?error=" + id.String())
				require.NoError(t, err)
				defer res.Body.Close()
				assert.EqualValues(t, http.StatusOK, res.StatusCode)

				actual, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)

				gg := make([]error, len(tc.gave))
				for k, g := range tc.gave {
					gg[k] = errorsx.Cause(g)
				}

				expected, err := json.Marshal(errorx.ErrorContainer{
					ID:     id,
					Errors: x.RequireJSONMarshal(t, gg),
				})
				require.NoError(t, err)

				assert.JSONEq(t, string(expected), string(actual), "%s != %s", expected, actual)
				assert.Empty(t, gjson.GetBytes(actual, "csrf_token").String())
				assert.JSONEq(t, string(x.RequireJSONMarshal(t, gg)), gjson.GetBytes(actual, "errors").Raw)
				t.Logf("%s", actual)
			})
		}
	})
}
