// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"

	"github.com/gofrs/uuid"
	otelattr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/x/otelx/semconv"
)

const (
	SessionIssued      semconv.Event = "SessionIssued"
	SessionRevoked     semconv.Event = "SessionRevoked"
	RegistrationFailed semconv.Event = "RegistrationFailed"
	IdentityCreated    semconv.Event = "IdentityCreated"
	LoginFailed        semconv.Event = "LoginFailed"
)

const (
	attributeKeyFlowType  semconv.AttributeKey = "FlowType"
	attributeKeySessionID semconv.AttributeKey = "SessionID"
)

func AttrFlowType(val flow.Type) otelattr.KeyValue {
	return otelattr.String(attributeKeyFlowType.String(), string(val))
}

func AttrSessionID(val uuid.UUID) otelattr.KeyValue {
	return otelattr.String(attributeKeySessionID.String(), val.String())
}

// Emit adds an event to the current span in the context.
func Emit(ctx context.Context, event semconv.Event, opt ...otelattr.KeyValue) {
	trace.SpanFromContext(ctx).AddEvent(
		event.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				opt...,
			)...,
		),
	)
}
