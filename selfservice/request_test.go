package selfservice

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/urlx"
)

func TestRequest(t *testing.T) {
	type i interface {
		Valid() error
		GetID() string
	}
	type f func(exp time.Duration) i

	for _, r := range []f{
		func(exp time.Duration) i {
			return NewLoginRequest(exp, &http.Request{URL: urlx.ParseOrPanic("/")})
		},
		func(exp time.Duration) i {
			return NewRegistrationRequest(exp, &http.Request{URL: urlx.ParseOrPanic("/")})
		},
	} {
		assert.NotEmpty(t, r(0).GetID())
		assert.NoError(t, r(time.Minute).Valid())
		assert.Error(t, r(-time.Minute).Valid())
	}
}
