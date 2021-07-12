package errorx_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ory/x/assertx"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/x"
	"github.com/ory/nosurf"
	"github.com/ory/x/errorsx"
)

func TestHandler(t *testing.T) {
	_, reg := internal.NewFastRegistryWithMocks(t)
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
		expectedError := x.MustEncodeJSON(t, herodot.ErrNotFound.WithReason("foobar"))

		t.Run("call with valid csrf cookie", func(t *testing.T) {
			hc := &http.Client{}
			id := getBody(t, hc, "/set-error", http.StatusOK)
			actual := getBody(t, hc, errorx.RouteGet+"?id="+string(id), http.StatusOK)
			assert.JSONEq(t, expectedError, gjson.GetBytes(actual, "error").Raw, "%s", actual)

			// We expect a forbid error if the error is not found, regardless of CSRF
			_ = getBody(t, hc, errorx.RouteGet+"?id=does-not-exist", http.StatusNotFound)
		})
	})

	t.Run("case=stubs", func(t *testing.T) {
		router := x.NewRouterPublic()
		h.RegisterPublicRoutes(router)
		ts := httptest.NewServer(router)
		defer ts.Close()

		res, err := ts.Client().Get(ts.URL + errorx.RouteGet + "?id=stub:500")
		require.NoError(t, err)
		require.EqualValues(t, http.StatusOK, res.StatusCode)

		actual, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)

		assert.EqualValues(t, "This is a stub error.", gjson.GetBytes(actual, "error.reason").String())
	})

	t.Run("case=errors types", func(t *testing.T) {
		router := x.NewRouterPublic()
		h.RegisterPublicRoutes(router)
		ts := httptest.NewServer(router)
		defer ts.Close()

		for k, tc := range []struct {
			gave error
		}{
			{gave: herodot.ErrNotFound.WithReason("foobar")},
			{gave: herodot.ErrNotFound.WithReason("foobar")},
			{gave: errors.WithStack(herodot.ErrNotFound.WithReason("foobar"))},
			{gave: errors.WithStack(herodot.ErrNotFound.WithReason("foobar").WithTrace(errors.New("asdf")))},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				csrf := x.NewUUID()
				id, err := reg.SelfServiceErrorPersister().Add(context.Background(), csrf.String(), tc.gave)
				require.NoError(t, err)

				res, err := ts.Client().Get(ts.URL + errorx.RouteGet + "?id=" + id.String())
				require.NoError(t, err)
				defer res.Body.Close()
				assert.EqualValues(t, http.StatusOK, res.StatusCode)

				actual, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)

				gg := errorsx.Cause(tc.gave)
				expected, err := json.Marshal(errorx.ErrorContainer{
					ID:     id,
					Errors: x.RequireJSONMarshal(t, gg),
				})
				require.NoError(t, err)

				assertx.EqualAsJSONExcept(t, json.RawMessage(expected), json.RawMessage(actual), []string{"created_at", "updated_at"})
				assert.Empty(t, gjson.GetBytes(actual, "csrf_token").String())
				assert.JSONEq(t, string(x.RequireJSONMarshal(t, gg)), gjson.GetBytes(actual, "error").Raw)
				t.Logf("%s", actual)
			})
		}
	})
}
