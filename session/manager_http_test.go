package session_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/session"
)

type mockCSRFHandler struct {
	c int
}

func (f *mockCSRFHandler) RegenerateToken(w http.ResponseWriter, r *http.Request) string {
	f.c++
	return "csrf_token"
}

func TestManagerHTTP(t *testing.T) {
	t.Run("method=SaveToRequest", func(t *testing.T) {
		_, reg := internal.NewRegistryDefault(t)

		mock := new(mockCSRFHandler)
		reg.WithCSRFHandler(mock)

		require.NoError(t, reg.SessionManager().SaveToRequest(context.Background(), new(session.Session), httptest.NewRecorder(), new(http.Request)))
		assert.Equal(t, 1, mock.c)
	})
}
