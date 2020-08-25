package flow

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestVerifyRequest(t *testing.T) {
	require.EqualError(t, VerifyRequest(&http.Request{}, TypeBrowser, x.FakeCSRFTokenGenerator, "not_csrf_token"), x.ErrInvalidCSRFToken.Error())
	require.NoError(t, VerifyRequest(&http.Request{}, TypeBrowser, x.FakeCSRFTokenGenerator, x.FakeCSRFToken), nil)
	require.NoError(t, VerifyRequest(&http.Request{}, TypeAPI, x.FakeCSRFTokenGenerator, ""))
	require.EqualError(t, VerifyRequest(&http.Request{
		Header: http.Header{"Origin":{"https://www.ory.sh"}},
	}, TypeAPI, x.FakeCSRFTokenGenerator, ""), ErrOriginHeaderNeedsBrowserFlow.Error())
	require.EqualError(t, VerifyRequest(&http.Request{
		Header: http.Header{"Cookie":{"cookie=ory"}},
	}, TypeAPI, x.FakeCSRFTokenGenerator, ""), ErrCookieHeaderNeedsBrowserFlow.Error())
}
