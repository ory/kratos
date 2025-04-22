// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"cmp"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.11.0"
	"go.opentelemetry.io/otel/trace"
	grpccodes "google.golang.org/grpc/codes"

	"github.com/ory/herodot"
	"github.com/ory/kratos/request"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx"
)

//go:embed password_migration_hook_default_mapper.jsonnet
var passwordHookDefaultMapper []byte

type (
	PasswordMigration struct {
		deps webHookDependencies
		conf *request.Config
	}
	PasswordMigrationRequest struct {
		templateContext
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	PasswordMigrationResponse struct {
		Status string `json:"status"`
	}
)

func NewPasswordMigrationHook(deps webHookDependencies, conf *request.Config) *PasswordMigration {
	return &PasswordMigration{deps: deps, conf: conf}
}

func (p *PasswordMigration) Execute(ctx context.Context, req *http.Request, flow flow.Flow, data *PasswordMigrationRequest) (err error) {
	var (
		httpClient = p.deps.HTTPClient(ctx)
		emitEvent  = p.conf.EmitAnalyticsEvent == nil || *p.conf.EmitAnalyticsEvent // default true
		tracer     = trace.SpanFromContext(ctx).TracerProvider().Tracer("kratos-webhooks")
	)
	p.conf.TemplateURI = cmp.Or(p.conf.TemplateURI, "base64://"+base64.StdEncoding.EncodeToString(passwordHookDefaultMapper))

	ctx, span := tracer.Start(ctx, "selfservice.login.password_migration")
	defer otelx.End(span, &err)

	if emitEvent {
		instrumentHTTPClientForEvents(ctx, httpClient, x.NewUUID(), "password_migration_hook")
	}
	builder, err := request.NewBuilder(ctx, p.conf, p.deps)
	if err != nil {
		return errors.WithStack(err)
	}
	data.templateContext = templateContext{
		Flow:           flow,
		RequestHeaders: req.Header,
		RequestMethod:  req.Method,
		RequestURL:     x.RequestURL(req).String(),
		RequestCookies: cookies(req),
	}
	whReq, err := builder.BuildRequest(ctx, data)
	if err != nil {
		return errors.WithStack(err)
	}
	rawData, err := json.Marshal(data)
	if err != nil {
		return errors.WithStack(err)
	}
	if err = whReq.SetBody(rawData); err != nil {
		return errors.WithStack(err)
	}

	p.deps.Logger().WithRequest(whReq.Request).Info("Dispatching password migration hook")
	whReq = whReq.WithContext(ctx)

	resp, err := httpClient.Do(whReq)
	if err != nil {
		return herodot.DefaultError{
			CodeField:     http.StatusBadGateway,
			StatusField:   http.StatusText(http.StatusBadGateway),
			GRPCCodeField: grpccodes.Aborted,
			ReasonField:   "A third-party upstream service could not be reached. Please try again later.",
			ErrorField:    "calling the password migration hook failed",
		}.WithWrap(errors.WithStack(err))
	}
	defer resp.Body.Close()
	span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(resp.StatusCode)...)

	switch resp.StatusCode {
	case http.StatusOK:
		// We now check if the response matches `{"status": "password_match" }`.
		dec := json.NewDecoder(io.LimitReader(resp.Body, 1024)) // limit the response body to 1KB
		var response PasswordMigrationResponse
		if err := dec.Decode(&response); err != nil || response.Status != "password_match" {
			return errors.WithStack(schema.NewInvalidCredentialsError())
		}
		return nil

	case http.StatusForbidden:
		return errors.WithStack(schema.NewInvalidCredentialsError())
	default:
		span.SetStatus(codes.Error, "Unexpected HTTP status code")
		return herodot.DefaultError{
			CodeField:     http.StatusBadGateway,
			StatusField:   http.StatusText(http.StatusBadGateway),
			GRPCCodeField: grpccodes.Aborted,
			ReasonField:   "A third-party upstream service responded improperly. Please try again later.",
			ErrorField:    fmt.Sprintf("password migration hook failed with status code %v", resp.StatusCode),
		}
	}
}
