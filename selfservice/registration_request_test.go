package selfservice

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/urlx"
)

func TestRegistrationRequest(t *testing.T) {
	r := NewRegistrationRequest(0, &http.Request{URL: urlx.ParseOrPanic("/")})
	assert.NotEmpty(t, r.ID)

	assert.NoError(t, NewRegistrationRequest(time.Minute, &http.Request{URL: urlx.ParseOrPanic("/")}).Valid())
	assert.Error(t, NewRegistrationRequest(-time.Minute, &http.Request{URL: urlx.ParseOrPanic("/")}).Valid())
}
