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
	CookieManager(ctx context.Context) sessions.StoreExact
	ContinuityCookieManager(ctx context.Context) sessions.StoreExact
}

type TracingProvider interface {
	Tracer(ctx context.Context) *tracing.Tracer
}

type SimpleLogger struct {
	L *logrusx.Logger
}

func (s *SimpleLogger) Logger() *logrusx.Logger {
	return s.L
}

func (s *SimpleLogger) Audit() *logrusx.Logger {
	return s.L
}

var _ LoggingProvider = (*SimpleLogger)(nil)
