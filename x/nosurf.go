package x

import (
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
)

type CSRFProvider interface {
	CSRFHandler() *nosurf.CSRFHandler
}

func NewCSRFHandler(router http.Handler, writer herodot.Writer) *nosurf.CSRFHandler {
	n := nosurf.New(router)
	n.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer.WriteError(w, r, errors.WithStack(herodot.ErrBadRequest.WithReasonf("CSRF token is missing or invalid.")))
	}))
	return n
}
