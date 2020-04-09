package password_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func newErrTs(t *testing.T, reg interface {
	errorx.PersistenceProvider
	x.WriterProvider
}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.SelfServiceErrorPersister().Read(r.Context(), x.ParseUUID(r.URL.Query().Get("error")))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e.Errors)
	}))
}

func newReturnTs(t *testing.T, reg interface {
	session.ManagementProvider
	x.WriterProvider
}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
		require.NoError(t, err)
		reg.Writer().Write(w, r, sess)
	}))
}

func hookConfig(u string) (m []map[string]interface{}) {
	var b bytes.Buffer
	if _, err := fmt.Fprintf(&b, `[
	{
		"job": "session"
	},
	{
		"job": "redirect",
		"config": {
          "default_redirect_url": "%s",
          "allow_user_defined_redirect": true
		}
	}
]`, u); err != nil {
		panic(err)
	}

	if err := json.NewDecoder(&b).Decode(&m); err != nil {
		panic(err)
	}

	return m
}
