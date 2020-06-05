package x

import (
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/pkg/errors"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/stringsx"

	"github.com/ory/herodot"
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

const FakeCSRFToken = "nosurf"

func FakeCSRFTokenGenerator(r *http.Request) string {
	return FakeCSRFToken
}

func FakeCSRFTokenGeneratorWithToken(token string) func(r *http.Request) string {
	return func(r *http.Request) string {
		return token
	}
}

type FakeCSRFHandler struct{ name string }

func NewFakeCSRFHandler(name string) *FakeCSRFHandler {
	return &FakeCSRFHandler{
		name: name,
	}
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
}

func NewCSRFHandler(
	router http.Handler,
	writer herodot.Writer,
	logger *logrusx.Logger,
	path string,
	domain string,
	secure bool,
) *nosurf.CSRFHandler {
	n := nosurf.New(router)
	n.SetBaseCookie(http.Cookie{
		MaxAge:   nosurf.MaxAge,
		Path:     path,
		Domain:   domain,
		HttpOnly: true,
		Secure:   secure,
	})
	n.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.
			WithField("expected_token", nosurf.Token(r)).
			WithField("received_token", r.Form.Get("csrf_token")).
			WithField("received_token_form", r.PostForm.Get("csrf_token")).
			Warn("A request failed due to a missing or invalid csrf_token value")

		writer.WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("CSRF token is missing or invalid.")))
	}))
	return n
}

func NewTestCSRFHandler(router http.Handler, reg interface {
	WithCSRFHandler(CSRFHandler)
	WithCSRFTokenGenerator(CSRFToken)
}) *nosurf.CSRFHandler {
	n := nosurf.New(router)
	n.SetBaseCookie(http.Cookie{MaxAge: nosurf.MaxAge, Path: "/", HttpOnly: true, Secure: false})
	reg.WithCSRFHandler(n)
	reg.WithCSRFTokenGenerator(nosurf.Token)
	return n
}
