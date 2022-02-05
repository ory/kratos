package flow

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/nosurf"
)

func TestGetCSRFToken(t *testing.T) {
	noToken := &mockReg{
		presentToken:     "",
		regeneratedToken: "regenerated",
	}

	tokenPresent := &mockReg{
		presentToken:     "existing",
		regeneratedToken: "regenerated",
	}

	t.Run("case=no token, browser flow", func(t *testing.T) {
		assert.Equal(t, "regenerated", GetCSRFToken(noToken, nil, nil, TypeBrowser))
	})

	t.Run("case=token present, browser flow", func(t *testing.T) {
		assert.Equal(t, "existing", GetCSRFToken(tokenPresent, nil, nil, TypeBrowser))
	})

	t.Run("case=no token, api flow", func(t *testing.T) {
		assert.Equal(t, "", GetCSRFToken(noToken, nil, nil, TypeAPI))
	})

	t.Run("case=token present, api flow", func(t *testing.T) {
		assert.Equal(t, "existing", GetCSRFToken(tokenPresent, nil, nil, TypeAPI))
	})
}

type mockReg struct {
	presentToken, regeneratedToken string

	nosurf.Handler
}

func (m *mockReg) GenerateCSRFToken(*http.Request) string {
	return m.presentToken
}

func (m *mockReg) CSRFHandler() nosurf.Handler {
	return m
}

func (m *mockReg) RegenerateToken(http.ResponseWriter, *http.Request) string {
	return m.regeneratedToken
}
