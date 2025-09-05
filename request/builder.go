// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package request

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/google/go-jsonnet"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/x"
	"github.com/ory/x/fetcher"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/otelx"

	"go.opentelemetry.io/otel/attribute"
)

var ErrCancel = errors.New("request cancel by JsonNet")

const (
	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeJSON = "application/json"
)

type (
	Dependencies interface {
		x.LoggingProvider
		x.TracingProvider
		x.HTTPClientProvider
		jsonnetsecure.VMProvider
	}
	Builder struct {
		r            *retryablehttp.Request
		Config       *Config
		deps         Dependencies
		cache        *ristretto.Cache[[]byte, []byte]
		bodySizeHint uint
	}
	options struct {
		cache        *ristretto.Cache[[]byte, []byte]
		bodySizeHint uint
	}
	BuilderOption = func(*options)
)

func WithCache(cache *ristretto.Cache[[]byte, []byte]) BuilderOption {
	return func(o *options) {
		o.cache = cache
	}
}

func WithBodySizeHint(hint uint) BuilderOption {
	return func(o *options) {
		o.bodySizeHint = hint
	}
}

func NewBuilder(ctx context.Context, c *Config, deps Dependencies, o ...BuilderOption) (_ *Builder, err error) {
	_, span := deps.Tracer(ctx).Tracer().Start(ctx, "request.NewBuilder")
	defer otelx.End(span, &err)

	var opts options
	for _, f := range o {
		f(&opts)
	}

	span.SetAttributes(
		attribute.String("url", c.URL),
		attribute.String("method", c.Method),
	)

	r, err := retryablehttp.NewRequest(c.Method, c.URL, nil)
	if err != nil {
		return nil, err
	}

	c.header = make(http.Header, len(c.Headers))
	for k, v := range c.Headers {
		c.header.Add(k, v)
	}
	if c.header.Get("Content-Type") == "" {
		c.header.Set("Content-Type", ContentTypeJSON)
	}

	c.auth, err = authStrategy(c.Auth.Type, c.Auth.Config)
	if err != nil {
		return nil, err
	}

	return &Builder{
		r:            r,
		Config:       c,
		deps:         deps,
		cache:        opts.cache,
		bodySizeHint: opts.bodySizeHint,
	}, nil
}

func (b *Builder) addBody(ctx context.Context, body interface{}) (err error) {
	ctx, span := b.deps.Tracer(ctx).Tracer().Start(ctx, "request.Builder.addBody")
	defer otelx.End(span, &err)

	if isNilInterface(body) {
		return nil
	}

	if b.Config.TemplateURI == "" {
		return errors.New("got empty template path for request with body")
	}

	tpl, err := b.readTemplate(ctx)
	if err != nil {
		return err
	}

	switch b.r.Header.Get("Content-Type") {
	case ContentTypeForm:
		if err := b.addURLEncodedBody(ctx, tpl, body); err != nil {
			return err
		}
	case "":
		b.r.Header.Set("Content-Type", ContentTypeJSON)
		fallthrough
	case ContentTypeJSON:
		if err := b.addJSONBody(ctx, tpl, body); err != nil {
			return err
		}
	default:
		return errors.New("invalid config - incorrect Content-Type for request with body")
	}

	return nil
}

func (b *Builder) addJSONBody(ctx context.Context, jsonnetSnippet []byte, body interface{}) error {
	buf := bytes.NewBuffer(make([]byte, 0, b.bodySizeHint))
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	if err := enc.Encode(body); err != nil {
		return errors.WithStack(err)
	}

	vm, err := b.deps.JsonnetVM(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	vm.TLACode("ctx", buf.String())

	res, err := vm.EvaluateAnonymousSnippet(
		b.Config.TemplateURI,
		string(jsonnetSnippet),
	)
	if err != nil {
		// Unfortunately we can not use errors.As / errors.Is, see:
		// https://github.com/google/go-jsonnet/issues/592
		if strings.Contains(err.Error(), (&jsonnet.RuntimeError{Msg: "cancel"}).Error()) {
			return errors.WithStack(ErrCancel)
		}

		return errors.WithStack(err)
	}

	rb := strings.NewReader(res)
	if err := b.r.SetBody(io.NopCloser(rb)); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (b *Builder) addURLEncodedBody(ctx context.Context, jsonnetSnippet []byte, body interface{}) error {
	buf := bytes.NewBuffer(make([]byte, 0, b.bodySizeHint))
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	if err := enc.Encode(body); err != nil {
		return errors.WithStack(err)
	}

	vm, err := b.deps.JsonnetVM(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	vm.TLACode("ctx", buf.String())

	res, err := vm.EvaluateAnonymousSnippet(b.Config.TemplateURI, string(jsonnetSnippet))
	if err != nil {
		return errors.WithStack(err)
	}

	values := map[string]string{}
	if err := json.Unmarshal([]byte(res), &values); err != nil {
		return errors.WithStack(err)
	}

	u := url.Values{}

	for key, value := range values {
		u.Add(key, value)
	}

	rb := strings.NewReader(u.Encode())
	if err := b.r.SetBody(io.NopCloser(rb)); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (b *Builder) BuildRequest(ctx context.Context, body interface{}) (*retryablehttp.Request, error) {
	b.r.Header = b.Config.header
	b.Config.auth.apply(b.r)

	// According to the HTTP spec any request method, but TRACE is allowed to
	// have a body. Even this is a bad practice for some of them, like for GET
	if b.Config.Method != http.MethodTrace {
		if err := b.addBody(ctx, body); err != nil {
			return nil, err
		}
	}

	return b.r, nil
}

func (b *Builder) addRawBody(body any) (err error) {
	if isNilInterface(body) {
		return nil
	}
	buf := bytes.NewBuffer(make([]byte, 0, b.bodySizeHint))
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(body); err != nil {
		return errors.WithStack(err)
	}
	switch contentType := b.r.Header.Get("Content-Type"); contentType {
	case "":
		b.r.Header.Set("Content-Type", ContentTypeJSON)
		fallthrough
	case ContentTypeJSON:
		if err := b.r.SetBody(buf); err != nil {
			return errors.WithStack(err)
		}
	default:
		return herodot.ErrMisconfiguration.WithDetail("invalid_content_type", contentType)
	}

	return nil
}

func (b *Builder) BuildRawRequest(body any) (*retryablehttp.Request, error) {
	b.r.Header = b.Config.header
	b.Config.auth.apply(b.r)

	// According to the HTTP spec any request method, but TRACE is allowed to
	// have a body. Even this is a bad practice for some of them, like for GET
	if b.Config.Method != http.MethodTrace {
		if err := b.addRawBody(body); err != nil {
			return nil, err
		}
	}

	return b.r, nil
}

func (b *Builder) readTemplate(ctx context.Context) ([]byte, error) {
	templateURI := b.Config.TemplateURI

	if templateURI == "" {
		return nil, nil
	}

	f := fetcher.NewFetcher(fetcher.WithClient(b.deps.HTTPClient(ctx)), fetcher.WithCache(b.cache, 60*time.Minute))
	tpl, err := f.FetchContext(ctx, templateURI)
	if errors.Is(err, fetcher.ErrUnknownScheme) {
		// legacy filepath
		templateURI = "file://" + templateURI
		b.deps.Logger().WithError(err).Warnf(
			"support for filepaths without a 'file://' scheme will be dropped in the next release, please use %s instead in your config",
			templateURI)

		tpl, err = f.FetchContext(ctx, templateURI)
	}
	// this handles the first error if it is a known scheme error, or the second fetch error
	if err != nil {
		return nil, err
	}

	return tpl.Bytes(), nil
}

func isNilInterface(i interface{}) bool {
	return i == nil || (reflect.ValueOf(i).Kind() == reflect.Pointer && reflect.ValueOf(i).IsNil())
}
