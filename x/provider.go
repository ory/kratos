package x

import (
	"context"

	"github.com/ory/x/httpx"
	"github.com/ory/x/tracing"

	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-retryablehttp"

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

type SimpleLoggerWithClient struct {
	L *logrusx.Logger
	C *retryablehttp.Client
}

func (s *SimpleLoggerWithClient) Logger() *logrusx.Logger {
	return s.L
}

func (s *SimpleLoggerWithClient) Audit() *logrusx.Logger {
	return s.L
}

func (s *SimpleLoggerWithClient) HTTPClient() *retryablehttp.Client {
	return s.C
}

var _ LoggingProvider = (*SimpleLoggerWithClient)(nil)
var _ HTTPClientProvider = (*SimpleLoggerWithClient)(nil)

type HTTPClientProvider interface {
	HTTPClient() *retryablehttp.Client
}

type ResilientHttpClient struct {
	client *retryablehttp.Client
}

func (r *ResilientHttpClient) HTTPClient() *retryablehttp.Client {
	return r.client
}

func NewResilientHttpClient(logger *logrusx.Logger) *ResilientHttpClient {
	return &ResilientHttpClient{
		client: httpx.NewResilientClient(httpx.ResilientClientWithLogger(logger)),
	}
}
