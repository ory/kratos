// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"context"

	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/x/httpx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
)

type CookieProvider interface {
	CookieManager(ctx context.Context) sessions.StoreExact
	ContinuityCookieManager(ctx context.Context) sessions.StoreExact
}

type BasicRegistry struct {
	L *logrusx.Logger
	C *retryablehttp.Client
	T *otelx.Tracer
}

func (s *BasicRegistry) Tracer(_ context.Context) *otelx.Tracer { return s.T }
func (s *BasicRegistry) Logger() *logrusx.Logger                { return s.L }

func (s *BasicRegistry) HTTPClient(_ context.Context, _ ...httpx.ResilientOptions) *retryablehttp.Client {
	return s.C
}

var _ logrusx.Provider = (*BasicRegistry)(nil)
var _ httpx.ClientProvider = (*BasicRegistry)(nil)
var _ otelx.Provider = (*BasicRegistry)(nil)
