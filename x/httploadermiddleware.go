package x

import (
	"context"
	"net/http"

	"github.com/urfave/negroni"
	"golang.org/x/oauth2"

	"github.com/ory/jsonschema/v3/httploader"
)

func HTTPLoaderContextMiddleware(reg interface {
	HTTPClientProvider
}) negroni.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		hc := reg.HTTPClient(r.Context())
		ctx := context.WithValue(r.Context(), oauth2.HTTPClient, hc)
		ctx = context.WithValue(ctx, httploader.ContextKey, hc)
		next(rw, r.WithContext(ctx))
	}
}
