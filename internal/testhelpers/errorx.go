package testhelpers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/logrusx"

	"github.com/ory/herodot"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func NewErrorTestServer(t *testing.T, reg interface {
	errorx.PersistenceProvider
	Configuration() *config.Config
}) *httptest.Server {
	logger := logrusx.New("", "", logrusx.ForceLevel(logrus.TraceLevel))
	writer := herodot.NewJSONWriter(logger)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.SelfServiceErrorPersister().Read(r.Context(), x.ParseUUID(r.URL.Query().Get("error")))
		require.NoError(t, err)
		t.Logf("Found error in NewErrorTestServer: %s", e.Errors)
		writer.Write(w, r, e.Errors)
	}))
	t.Cleanup(ts.Close)
	reg.Configuration().MustSet(config.ViperKeySelfServiceErrorUI, ts.URL)
	return ts
}

func NewRedirTS(t *testing.T, body string, conf *config.Config) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(body) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(ts.Close)
	conf.MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, ts.URL)
	return ts
}

func NewRedirSessionEchoTS(t *testing.T, reg interface {
	x.WriterProvider
	session.ManagementProvider
	Configuration() *config.Config
}) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
		require.NoError(t, err, "Headers: %+v", r.Header)
		reg.Writer().Write(w, r, sess)
	}))
	t.Cleanup(ts.Close)
	reg.Configuration().MustSet(config.ViperKeySelfServiceBrowserDefaultReturnTo, ts.URL+"/return-ts")
	return ts
}
