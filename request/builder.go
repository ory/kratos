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

	"github.com/dgraph-io/ristretto"
	"github.com/google/go-jsonnet"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ory/kratos/x"
	"github.com/ory/x/fetcher"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/otelx"
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
		r      *retryablehttp.Request
		Config *Config
		deps   Dependencies
		cache  *ristretto.Cache[[]byte, []byte]
	}
	options struct {
		cache *ristretto.Cache[[]byte, []byte]
	}
	BuilderOption = func(*options)
)

func WithCache(cache *ristretto.Cache[[]byte, []byte]) BuilderOption {
	return func(o *options) {
		o.cache = cache
	}
}

func NewBuilder(ctx context.Context, config json.RawMessage, deps Dependencies, o ...BuilderOption) (_ *Builder, err error) {
	_, span := deps.Tracer(ctx).Tracer().Start(ctx, "request.NewBuilder")
	defer otelx.End(span, &err)

	var opts options
	for _, f := range o {
		f(&opts)
	}

	c := Config{}
	if err := json.Unmarshal(config, &c); err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.String("url", c.URL),
		attribute.String("method", c.Method),
	)

	r, err := retryablehttp.NewRequest(c.Method, c.URL, nil)
	if err != nil {
		return nil, err
	}

	return &Builder{
		r:      r,
		Config: &c,
		deps:   deps,
		cache:  opts.cache,
	}, nil
}

func (b *Builder) addAuth() error {
	authConfig := b.Config.Auth

	strategy, err := authStrategy(authConfig.Type, authConfig.Config)
	if err != nil {
		return err
	}

	strategy.apply(b.r)

	return nil
}

func (b *Builder) addBody(ctx context.Context, body interface{}) (err error) {
	ctx, span := b.deps.Tracer(ctx).Tracer().Start(ctx, "request.Builder.addBody")
	defer otelx.End(span, &err)

	if isNilInterface(body) {
		return nil
	}

	contentType := b.r.Header.Get("Content-Type")

	if b.Config.TemplateURI == "" {
		return errors.New("got empty template path for request with body")
	}

	tpl, err := b.readTemplate(ctx)
	if err != nil {
		return err
	}

	switch contentType {
	case ContentTypeForm:
		if err := b.addURLEncodedBody(ctx, tpl, body); err != nil {
			return err
		}
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
	buf := new(bytes.Buffer)
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
	buf := new(bytes.Buffer)
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
	b.r.Header = b.Config.Header
	if err := b.addAuth(); err != nil {
		return nil, err
	}

	// According to the HTTP spec any request method, but TRACE is allowed to
	// have a body. Even this is a bad practice for some of them, like for GET
	if b.Config.Method != http.MethodTrace {
		if err := b.addBody(ctx, body); err != nil {
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
	return i == nil || (reflect.ValueOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil())
}
