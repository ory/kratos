package x

import (
	"encoding/base64"
	"net/http"

	"github.com/ory/kratos/driver/config"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/nosurf"
	"github.com/ory/x/randx"
	"github.com/ory/x/stringsx"
)

var (
	ErrInvalidCSRFToken = herodot.ErrForbidden.WithReasonf("A request failed due to a missing or invalid csrf_token value.")
	ErrGone             = herodot.DefaultError{
		CodeField:    http.StatusGone,
		StatusField:  http.StatusText(http.StatusGone),
		ReasonField:  "",
		DebugField:   "",
		DetailsField: nil,
		ErrorField:   "The requested resource is no longer available because it has expired or is otherwise invalid.",
	}
)

type CSRFTokenGeneratorProvider interface {
	GenerateCSRFToken(r *http.Request) string
}

type CSRFToken func(r *http.Request) string

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

var _ CSRFHandler = new(FakeCSRFHandler)

type FakeCSRFHandler struct{ name string }

func NewFakeCSRFHandler(name string) *FakeCSRFHandler {
	return &FakeCSRFHandler{
		name: name,
	}
}

func (f *FakeCSRFHandler) ExemptPath(s string) {
}

func (f *FakeCSRFHandler) IgnorePath(s string) {
}

func (f *FakeCSRFHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (f *FakeCSRFHandler) RegenerateToken(w http.ResponseWriter, r *http.Request) string {
	return stringsx.Coalesce(f.name, FakeCSRFToken)
}

type CSRFProvider interface {
	CSRFHandler() CSRFHandler
}

type CSRFHandler interface {
	http.Handler
	RegenerateToken(w http.ResponseWriter, r *http.Request) string
	ExemptPath(string)
	IgnorePath(string)
}

func NosurfBaseCookieHandler(reg interface {
	config.Provider
}) func(w http.ResponseWriter, r *http.Request) http.Cookie {
	return func(w http.ResponseWriter, r *http.Request) http.Cookie {
		secure := !reg.Config(r.Context()).IsInsecureDevMode()

		sameSite := http.SameSiteNoneMode
		if !secure {
			sameSite = http.SameSiteLaxMode
		}

		name := base64.RawURLEncoding.EncodeToString([]byte(reg.Config(r.Context()).SelfPublicURL().String())) + "_csrf_token"

		return http.Cookie{
			Name:     name,
			MaxAge:   nosurf.MaxAge,
			Path:     stringsx.Coalesce(reg.Config(r.Context()).SelfPublicURL().Path, "/"),
			Domain:   reg.Config(r.Context()).SelfPublicURL().Hostname(),
			HttpOnly: true,
			Secure:   secure,
			SameSite: sameSite,
		}
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
	n.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		reg.Logger().
			WithField("result", nosurf.VerifyToken(nosurf.Token(r), r.Form.Get("csrf_token"))).
			WithField("expected_token", nosurf.Token(r)).
			WithField("received_cookies", r.Cookies()).
			WithField("received_token_form", r.Form.Get("csrf_token")).
			WithField("received_token_body", r.PostForm.Get("csrf_token")).
			Warn("A request failed due to a missing or invalid csrf_token value")

		reg.Writer().WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("CSRF token is missing or invalid.")))
	}))
	return n
}

func NewTestCSRFHandler(router http.Handler, reg interface {
	WithCSRFHandler(CSRFHandler)
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
