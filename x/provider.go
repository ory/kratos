// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"context"

	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
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
	Tracer(ctx context.Context) *otelx.Tracer
}

type SimpleLoggerWithClient struct {
	L *logrusx.Logger
	C *retryablehttp.Client
	T *otelx.Tracer
}

func (s *SimpleLoggerWithClient) Tracer(_ context.Context) *otelx.Tracer {
	return s.T
}

func (s *SimpleLoggerWithClient) Logger() *logrusx.Logger {
	return s.L
}

func (s *SimpleLoggerWithClient) Audit() *logrusx.Logger {
	return s.L
}

func (s *SimpleLoggerWithClient) HTTPClient(_ context.Context, _ ...httpx.ResilientOptions) *retryablehttp.Client {
	return s.C
}

var _ LoggingProvider = (*SimpleLoggerWithClient)(nil)
var _ HTTPClientProvider = (*SimpleLoggerWithClient)(nil)
var _ TracingProvider = (*SimpleLoggerWithClient)(nil)
