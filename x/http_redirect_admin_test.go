package x

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"
)

func TestRedirectAdmin(t *testing.T) {
	router := httprouter.New()
	router.GET("/admin/identities", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = w.Write([]byte("identities"))
	})
	n := negroni.New()
	n.UseFunc(RedirectAdminMiddleware)
	n.UseHandler(router)
	ts := httptest.NewServer(n)
	defer ts.Close()

	for _, tc := range []struct {
		loc          string
		expectedCode int
		expectedBody string
		expectedPath string
	}{
		{loc: "/identities", expectedCode: 200, expectedBody: "identities", expectedPath: "/admin/identities"},
		{loc: "/admin/identities", expectedCode: 200, expectedBody: "identities", expectedPath: "/admin/identities"},
		{loc: "/not/exist", expectedCode: 404, expectedPath: "/admin/not/exist"},
		{loc: "/", expectedCode: 404, expectedPath: "/admin"},
		{loc: "/admin", expectedCode: 404, expectedPath: "/admin"},
		{loc: "/admin/", expectedCode: 404, expectedPath: "/admin/"},
		{loc: "", expectedCode: 404, expectedPath: "/admin"},
	} {
		t.Run(tc.loc, func(t *testing.T) {
			res, err := ts.Client().Get(ts.URL + tc.loc)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			assert.Equal(t, tc.expectedPath, res.Request.URL.Path)
			defer res.Body.Close()
			if tc.expectedBody != "" {
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedBody, strings.TrimSpace(string(body)))
			}
		})
	}
}
