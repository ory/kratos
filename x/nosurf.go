package x

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/ory/kratos/text"

	"github.com/ory/kratos/driver/config"

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

func FakeCSRFTokenGenerator(r *http.Request) string {
	return FakeCSRFToken
}

func FakeCSRFTokenGeneratorWithToken(token string) func(r *http.Request) string {
	return func(r *http.Request) string {
		return token
	}
}

var _ nosurf.Handler = new(FakeCSRFHandler)

type FakeCSRFHandler struct{ name string }

func NewFakeCSRFHandler(name string) *FakeCSRFHandler {
	return &FakeCSRFHandler{
		name: name,
	}
}

func (f *FakeCSRFHandler) DisablePath(s string) {
}

func (f *FakeCSRFHandler) DisableGlob(s string) {
}

func (f *FakeCSRFHandler) DisableGlobs(s ...string) {
}

func (f *FakeCSRFHandler) ExemptPath(s string) {
}

func (f *FakeCSRFHandler) IgnorePath(s string) {
}

func (f *FakeCSRFHandler) IgnoreGlob(s string) {
}

func (f *FakeCSRFHandler) IgnoreGlobs(s ...string) {
}

func (f *FakeCSRFHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (f *FakeCSRFHandler) RegenerateToken(w http.ResponseWriter, r *http.Request) string {
	return stringsx.Coalesce(f.name, FakeCSRFToken)
}

type CSRFProvider interface {
	CSRFHandler() nosurf.Handler
}

func CSRFCookieName(reg interface {
	config.Provider
}, r *http.Request) string {
	return "csrf_token_" + fmt.Sprintf("%x", sha256.Sum256([]byte(reg.Config(r.Context()).SelfPublicURL().String())))
}

func NosurfBaseCookieHandler(reg interface {
	config.Provider
}) func(w http.ResponseWriter, r *http.Request) http.Cookie {
	return func(w http.ResponseWriter, r *http.Request) http.Cookie {
		secure := !reg.Config(r.Context()).IsInsecureDevMode()

		sameSite := reg.Config(r.Context()).CookieSameSiteMode()
		if !secure {
			sameSite = http.SameSiteLaxMode
		}

		domain := ""
		if d := reg.Config(r.Context()).CookieDomain(); d != "" {
			domain = d
		}

		name := CSRFCookieName(reg, r)
		cookie := http.Cookie{
			Name:     name,
			MaxAge:   nosurf.MaxAge,
			Path:     reg.Config(r.Context()).CookiePath(),
			Domain:   domain,
			HttpOnly: true,
			Secure:   secure,
			SameSite: sameSite,
		}

		if alias := reg.Config(r.Context()).SelfPublicURL(); reg.Config(r.Context()).SelfPublicURL().String() != alias.String() {
			// If a domain alias is detected use that instead.
			cookie.Domain = alias.Hostname()
			cookie.Path = alias.Path
		}

		return cookie
	}
}

func CSRFErrorReason(r *http.Request, reg interface {
	config.Provider
}) error {
	// Is it an AJAX request?
	isAjax := len(r.Header.Get("Origin")) == 0

	if len(r.Header.Get("Cookie")) == 0 {
		if isAjax {
			return errors.WithStack(ErrInvalidCSRFTokenAJAXNoCookies)
		}
		return errors.WithStack(ErrInvalidCSRFTokenServerNoCookies)
	} else if _, err := r.Cookie(CSRFCookieName(reg, r)); errors.Is(err, http.ErrNoCookie) {
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
	config.Provider
	LoggingProvider
	WriterProvider
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := CSRFErrorReason(r, reg)
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
		config.Provider
		LoggingProvider
		WriterProvider
	}) *nosurf.CSRFHandler {
	n := nosurf.New(router)

	n.SetBaseCookieFunc(NosurfBaseCookieHandler(reg))
	n.SetFailureHandler(CSRFFailureHandler(reg))
	return n
}

func NewTestCSRFHandler(router http.Handler, reg interface {
	WithCSRFHandler(handler nosurf.Handler)
	WithCSRFTokenGenerator(CSRFToken)
	WriterProvider
	LoggingProvider
	config.Provider
}) *nosurf.CSRFHandler {
	n := NewCSRFHandler(router, reg)
	reg.WithCSRFHandler(n)
	reg.WithCSRFTokenGenerator(nosurf.Token)
	return n
}
