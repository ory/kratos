// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.11.0"
	"go.opentelemetry.io/otel/trace"
	grpccodes "google.golang.org/grpc/codes"

	"github.com/ory/herodot"
	"github.com/ory/kratos/request"
	"github.com/ory/kratos/schema"
	"github.com/ory/x/otelx"
)

type (
	PasswordMigration struct {
		deps webHookDependencies
		conf json.RawMessage
	}
	PasswordMigrationData struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
)

var passwordMigrationJsonnetTemplate string

func init() {
	snippet := []byte(`function(ctx) { password: ctx.password, identifier: ctx.identifier }`)
	passwordMigrationJsonnetTemplate = "base64://" + base64.StdEncoding.EncodeToString(snippet)
}

func NewPasswordMigrationHook(deps webHookDependencies, conf json.RawMessage) *PasswordMigration {
	return &PasswordMigration{deps: deps, conf: conf}
}

func (p *PasswordMigration) Execute(ctx context.Context, data *PasswordMigrationData) (err error) {
	var (
		httpClient = p.deps.HTTPClient(ctx)
		emitEvent  = gjson.GetBytes(p.conf, "emit_analytics_event").Bool() || !gjson.GetBytes(p.conf, "emit_analytics_event").Exists() // default true
		tracer     = trace.SpanFromContext(ctx).TracerProvider().Tracer("kratos-webhooks")
	)

	ctx, span := tracer.Start(ctx, "selfservice.login.password_migration")
	defer otelx.End(span, &err)

	if emitEvent {
		instrumentHTTPClientForEvents(ctx, httpClient)
	}
	builder, err := request.NewBuilder(ctx, p.conf, p.deps, jsonnetCache)
	if err != nil {
		return err
	}

	builder.Config.TemplateURI = passwordMigrationJsonnetTemplate

	req, err := builder.BuildRequest(ctx, data)
	if errors.Is(err, request.ErrCancel) {
		span.SetAttributes(attribute.Bool("password_migration.jsonnet.canceled", true))
		return nil
	} else if err != nil {
		return err
	}

	p.deps.Logger().WithRequest(req.Request).Info("Dispatching password migration hook")
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		if isTimeoutError(err) {
			return herodot.DefaultError{
				CodeField:     http.StatusGatewayTimeout,
				StatusField:   http.StatusText(http.StatusGatewayTimeout),
				GRPCCodeField: grpccodes.DeadlineExceeded,
				ErrorField:    err.Error(),
				ReasonField:   "A third-party upstream service could not be reached. Please try again later.",
			}.WithWrap(errors.WithStack(err))
		}
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
