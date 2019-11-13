package errorx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
)

func NewErrorTestServer(t *testing.T, reg ManagementProvider) *httptest.Server {
	logger := logrus.New()
	writer := herodot.NewJSONWriter(logger)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.ErrorManager().Read(r.Context(), r.URL.Query().Get("error"))
		require.NoError(t, err)
		logger.Errorf("Found error in NewErrorTestServer: %s", e)
		writer.Write(w, r, e)
	}))
}
