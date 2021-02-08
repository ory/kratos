package x_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
	"github.com/ory/nosurf"
	"github.com/ory/x/randx"
)

func TestNosurfBaseCookieHandler(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	require.NoError(t, conf.Source().Set(config.ViperKeyPublicBaseURL, "http://foo.com/bar"))

	cookie := x.NosurfBaseCookieHandler(reg)(httptest.NewRecorder(), httptest.NewRequest("GET", "https://foo/bar", nil))
	assert.EqualValues(t, "aHR0cDovL2Zvby5jb20vYmFy_csrf_token", cookie.Name, "base64 representation of http://foo.com/bar")
	assert.EqualValues(t, http.SameSiteLaxMode, cookie.SameSite, "is set to lax because https/secure is false - chrome rejects none samesite on non-https")
	assert.EqualValues(t, nosurf.MaxAge, cookie.MaxAge)
	assert.False(t, cookie.Secure, "false because insecure dev mode")
	assert.True(t, cookie.HttpOnly)

	for i := 0; i < 10; i++ {
		require.NoError(t, conf.Source().Set(config.ViperKeyPublicBaseURL, randx.MustString(16, randx.AlphaNum)))
		cookie := x.NosurfBaseCookieHandler(reg)(httptest.NewRecorder(), httptest.NewRequest("GET", "https://foo/bar", nil))

		assert.NotEqual(t, "aHR0cDovL2Zvby5jb20vYmFy_csrf_token", cookie.Name, "should no longer be http://foo.com/bar")

		matches, err := regexp.MatchString("[a-zA-Z_0-9]+", cookie.Name)
		require.NoError(t, err)
		assert.True(t, matches, "does not have any special chars")
	}

	require.NoError(t, conf.Source().Set("dev", false))
	cookie = x.NosurfBaseCookieHandler(reg)(httptest.NewRecorder(), httptest.NewRequest("GET", "https://foo/bar", nil))
	assert.EqualValues(t, http.SameSiteNoneMode, cookie.SameSite, "can be none because https/secure is true")
	assert.True(t, cookie.Secure, "true because secure mode")
	assert.True(t, cookie.HttpOnly)
}
