package x

import (
	"context"
	"github.com/ory/jsonschema/v3/httploader"
	"github.com/urfave/negroni"
	"net/http"
)

func HTTPLoaderContextMiddleware(reg interface {
	HTTPClientProvider
}) negroni.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(context.WithValue(r.Context(), httploader.ContextKey, reg.HTTPClient(r.Context()))))
	}
}
