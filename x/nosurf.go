package x

import (
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ory/herodot"
)

var (
	ErrInvalidCSRFToken = herodot.ErrForbidden.WithReasonf("A request failed due to a missing or invalid csrf_token value.")
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

type FakeCSRFHandler struct {}

func (f *FakeCSRFHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (f *FakeCSRFHandler) RegenerateToken(w http.ResponseWriter, r *http.Request) string {
	return FakeCSRFToken
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
	logger logrus.FieldLogger,
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

func NewTestCSRFHandler(router http.Handler) *nosurf.CSRFHandler {
	n := nosurf.New(router)
	n.SetBaseCookie(http.Cookie{
		MaxAge:   nosurf.MaxAge,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})
	return n
}
