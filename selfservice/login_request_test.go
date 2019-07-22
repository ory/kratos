package selfservice

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/urlx"
)

func TestLoginRequest(t *testing.T) {
	r := NewLoginRequest(0, &http.Request{URL: urlx.ParseOrPanic("/")})
	assert.NotEmpty(t, r.ID)

	assert.NoError(t, NewLoginRequest(time.Minute, &http.Request{URL: urlx.ParseOrPanic("/")}).Valid())
	assert.Error(t, NewLoginRequest(-time.Minute, &http.Request{URL: urlx.ParseOrPanic("/")}).Valid())
}
