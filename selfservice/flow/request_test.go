package flow

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestVerifyRequest(t *testing.T) {
	require.EqualError(t, EnsureCSRF(&http.Request{}, TypeBrowser, false, x.FakeCSRFTokenGenerator, "not_csrf_token"), x.ErrInvalidCSRFToken.Error())
	require.NoError(t, EnsureCSRF(&http.Request{}, TypeBrowser, false, x.FakeCSRFTokenGenerator, x.FakeCSRFToken), nil)
	require.NoError(t, EnsureCSRF(&http.Request{}, TypeAPI, false, x.FakeCSRFTokenGenerator, ""))
	require.EqualError(t, EnsureCSRF(&http.Request{
		Header: http.Header{"Origin": {"https://www.ory.sh"}},
	}, TypeAPI, false, x.FakeCSRFTokenGenerator, ""), ErrOriginHeaderNeedsBrowserFlow.Error())
	require.EqualError(t, EnsureCSRF(&http.Request{
		Header: http.Header{"Cookie": {"cookie=ory"}},
	}, TypeAPI, false, x.FakeCSRFTokenGenerator, ""), ErrCookieHeaderNeedsBrowserFlow.Error())
}
