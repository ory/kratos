package flow

import (
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/ory/herodot"
	"github.com/pkg/errors"

	"github.com/ory/kratos/x"
)

var ErrOriginHeaderNeedsBrowserFlow = herodot.ErrBadRequest.
	WithReasonf(`The HTTP Request Header included the "Origin" key, indicating that this request was made as part of an AJAX request in a Browser. The flow however was initiated as an API request. To prevent potential misuse and mitigate several attack vectors including CSRF, the request has been blocked. Please consult the documentation.`)
var ErrCookieHeaderNeedsBrowserFlow = herodot.ErrBadRequest.
	WithReasonf(`The HTTP Request Header included the "Cookie" key, indicating that this request was made by a Browser. The flow however was initiated as an API request. To prevent potential misuse and mitigate several attack vectors including CSRF, the request has been blocked. Please consult the documentation.`)

func VerifyRequest(
	r *http.Request,
	flowType Type,
	generator func(r *http.Request) string,
	actual string,
) error {
	switch flowType {
	case TypeAPI:
		// API Based flows to not require anti-CSRF tokens because we can not leverage a session, making this
		// endpoint pointless.

		// Let's ensure that no-one mistakenly makes an AJAX request using the API flow.
		if r.Header.Get("Origin") != "" {
			return errors.WithStack(ErrOriginHeaderNeedsBrowserFlow)
		}

		if len(r.Cookies()) > 0 {
			return errors.WithStack(ErrCookieHeaderNeedsBrowserFlow)
		}

		return nil
	default:
		if !nosurf.VerifyToken(generator(r), actual) {
			return x.ErrInvalidCSRFToken
		}
	}

	return nil
}
