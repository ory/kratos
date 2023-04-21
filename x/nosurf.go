// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ory/kratos/text"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/nosurf"
	"github.com/ory/x/randx"
	"github.com/ory/x/stringsx"
)

var (
	ErrInvalidCSRFToken = herodot.ErrForbidden.
				WithID(text.ErrIDCSRF).
				WithError("the request was rejected to protect you from Cross-Site-Request-Forgery").
				WithDetail("docs", "https://www.ory.sh/kratos/docs/debug/csrf").
				WithReason("Please retry the flow and optionally clear your cookies. The request was rejected to protect you from Cross-Site-Request-Forgery (CSRF) which could cause account takeover, leaking personal information, and other serious security issues.")
	ErrGone = herodot.DefaultError{
		CodeField:    http.StatusGone,
		StatusField:  http.StatusText(http.StatusGone),
		ReasonField:  "",
		DebugField:   "",
		DetailsField: nil,
		ErrorField:   "The requested resource is no longer available because it has expired or is otherwise invalid.",
	}
)

const noCookie = "The HTTP Cookie Header is empty or not set."
const cookieMissing = "The HTTP Cookie Header was set but did not include the anti-CSRF cookie."
const tokenNotSent = "The anti-CSRF cookie was found but the CSRF token was not included in the HTTP request body (" + nosurf.CookieName + ") nor in the HTTP Header (" + nosurf.HeaderName + ")."
const tokenMismatch = "The HTTP Cookie Header was set and a CSRF token was sent but they do not match. We recommend deleting all cookies for this domain and retrying the flow."

var (
	ErrInvalidCSRFTokenAJAX = ErrInvalidCSRFToken.
				WithDetail("hint", "We detected an AJAX call, please ensure that CORS is enabled and configured correctly, and that your AJAX code sends cookies and has credentials enabled. For further debugging, check your Browser's Network Tab to see what cookies are included or excluded.")

	ErrInvalidCSRFTokenAJAXNoCookies     = ErrInvalidCSRFTokenAJAX.WithDetail("reject_reason", noCookie)
	ErrInvalidCSRFTokenAJAXCookieMissing = ErrInvalidCSRFTokenAJAX.WithDetail("reject_reason", cookieMissing)
	ErrInvalidCSRFTokenAJAXTokenNotSent  = ErrInvalidCSRFToken.WithDetail("hint", tokenNotSent)
	ErrInvalidCSRFTokenAJAXTokenMismatch = ErrInvalidCSRFTokenAJAX.WithDetail("reject_reason", tokenMismatch)
)

var (
	ErrInvalidCSRFTokenServer = ErrInvalidCSRFToken.
					WithDetail("hint", "We detected a regular browser or server-side call. To debug browser calls check your Browser's Network Tab to see what cookies are included or excluded. If you are calling from a server ensure that the appropriate cookies are being forwarded and that the SDK method is called correctly.")

	ErrInvalidCSRFTokenServerNoCookies     = ErrInvalidCSRFTokenServer.WithDetail("reject_reason", noCookie)
	ErrInvalidCSRFTokenServerCookieMissing = ErrInvalidCSRFTokenServer.WithDetail("reject_reason", cookieMissing)
	ErrInvalidCSRFTokenServerTokenNotSent  = ErrInvalidCSRFToken.WithDetail("hint", tokenNotSent)
	ErrInvalidCSRFTokenServerTokenMismatch = ErrInvalidCSRFTokenAJAX.WithDetail("reject_reason", tokenMismatch)
)

type CSRFTokenGeneratorProvider interface {
	GenerateCSRFToken(r *http.Request) string
}

type CSRFToken func(r *http.Request) string

const CSRFTokenName = "csrf_token"

func DefaultCSRFToken(r *http.Request) string {
	return nosurf.Token(r)
}

var FakeCSRFToken = base64.StdEncoding.EncodeToString([]byte(randx.MustString(32, randx.AlphaLowerNum)))

func FakeCSRFTokenGenerator(*http.Request) string {
	return FakeCSRFToken
}

var _ nosurf.Handler = new(FakeCSRFHandler)

type FakeCSRFHandler struct{ name string }

func NewFakeCSRFHandler(name string) *FakeCSRFHandler {
	return &FakeCSRFHandler{
		name: name,
	}
}

func (f *FakeCSRFHandler) DisablePath(string) {
}

func (f *FakeCSRFHandler) DisableGlob(string) {
}

func (f *FakeCSRFHandler) DisableGlobs(...string) {
}

func (f *FakeCSRFHandler) ExemptPath(string) {
}

func (f *FakeCSRFHandler) IgnorePath(string) {
}

func (f *FakeCSRFHandler) IgnoreGlob(string) {
}

func (f *FakeCSRFHandler) IgnoreGlobs(...string) {
}

func (f *FakeCSRFHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
}

func (f *FakeCSRFHandler) RegenerateToken(http.ResponseWriter, *http.Request) string {
	return stringsx.Coalesce(f.name, FakeCSRFToken)
}

type CSRFProvider interface {
	CSRFHandler() nosurf.Handler
}

func CSRFCookieName(c interface {
	SelfPublicURL(ctx context.Context) *url.URL
}, r *http.Request) string {
	return "csrf_token_" + fmt.Sprintf("%x", sha256.Sum256([]byte(c.SelfPublicURL(r.Context()).String())))
}

type config interface {
	SelfPublicURL(ctx context.Context) *url.URL
	IsInsecureDevMode(ctx context.Context) bool
	CookieSameSiteMode(ctx context.Context) http.SameSite
	CookiePath(ctx context.Context) string
	CookieDomain(ctx context.Context) string
}

func NosurfBaseCookieHandler(c config) func(w http.ResponseWriter, r *http.Request) http.Cookie {
	return func(w http.ResponseWriter, r *http.Request) http.Cookie {
		secure := !c.IsInsecureDevMode(r.Context())

		sameSite := c.CookieSameSiteMode(r.Context())
		if !secure {
			sameSite = http.SameSiteLaxMode
		}

		domain := ""
		if d := c.CookieDomain(r.Context()); d != "" {
			domain = d
		}

		name := CSRFCookieName(c, r)
		cookie := http.Cookie{
			Name:     name,
			MaxAge:   nosurf.MaxAge,
			Path:     c.CookiePath(r.Context()),
			Domain:   domain,
			HttpOnly: true,
			Secure:   secure,
			SameSite: sameSite,
		}

		return cookie
	}
}

func CSRFErrorReason(r *http.Request, c config) error {
	// Is it an AJAX request?
	isAjax := len(r.Header.Get("Origin")) == 0

	if len(r.Header.Get("Cookie")) == 0 {
		if isAjax {
			return errors.WithStack(ErrInvalidCSRFTokenAJAXNoCookies)
		}
		return errors.WithStack(ErrInvalidCSRFTokenServerNoCookies)
	} else if _, err := r.Cookie(CSRFCookieName(c, r)); errors.Is(err, http.ErrNoCookie) {
		if isAjax {
			return errors.WithStack(ErrInvalidCSRFTokenAJAXCookieMissing)
		}
		return errors.WithStack(ErrInvalidCSRFTokenServerCookieMissing)
	} else if len(r.Form.Get("csrf_token")+r.Header.Get(nosurf.HeaderName)) == 0 {
		if isAjax {
			return errors.WithStack(ErrInvalidCSRFTokenAJAXTokenNotSent)
		}
		return errors.WithStack(ErrInvalidCSRFTokenServerTokenNotSent)
	}

	if isAjax {
		return errors.WithStack(ErrInvalidCSRFTokenAJAXTokenMismatch)
	}
	return errors.WithStack(ErrInvalidCSRFTokenServerTokenMismatch)
}

func CSRFFailureHandler(reg interface {
	LoggingProvider
	WriterProvider
}, c config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := CSRFErrorReason(r, c)
		reg.Logger().
			WithError(err).
			WithField("result", nosurf.VerifyToken(nosurf.Token(r), r.Form.Get("csrf_token"))).
			WithField("expected_token", nosurf.Token(r)).
			WithField("received_cookies", r.Cookies()).
			WithField("received_token_form", r.Form.Get(nosurf.FormFieldName)).
			WithField("received_token_body", r.PostForm.Get(nosurf.FormFieldName)).
			WithField("received_token_header", r.Header.Get(nosurf.HeaderName)).
			Warn("A request failed due to a missing or invalid csrf_token value")

		reg.Writer().WriteError(w, r, err)
	}
}

func NewCSRFHandler(
	router http.Handler,
	reg interface {
		LoggingProvider
		WriterProvider
	},
	c config) *nosurf.CSRFHandler {
	n := nosurf.New(router)

	n.SetBaseCookieFunc(NosurfBaseCookieHandler(c))
	n.SetFailureHandler(CSRFFailureHandler(reg, c))
	return n
}

func NewTestCSRFHandler(router http.Handler, reg interface {
	WithCSRFHandler(handler nosurf.Handler)
	WithCSRFTokenGenerator(CSRFToken)
	WriterProvider
	LoggingProvider
}, c config) *nosurf.CSRFHandler {
	n := NewCSRFHandler(router, reg, c)
	reg.WithCSRFHandler(n)
	reg.WithCSRFTokenGenerator(nosurf.Token)
	return n
}
