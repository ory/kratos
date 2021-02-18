package x

import (
	"context"

	"github.com/ory/x/tracing"

	"github.com/gorilla/sessions"

	"github.com/ory/herodot"
	"github.com/ory/x/logrusx"
)

type LoggingProvider interface {
	Logger() *logrusx.Logger
	Audit() *logrusx.Logger
}

type WriterProvider interface {
	Writer() herodot.Writer
}

type CookieProvider interface {
	CookieManager(ctx context.Context) sessions.Store
	ContinuityCookieManager(ctx context.Context) sessions.Store
}

type TracingProvider interface {
	Tracer(ctx context.Context) *tracing.Tracer
}
