// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"reflect"

	"github.com/gofrs/uuid"
	otelattr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	f "github.com/ory/kratos/selfservice/flow"

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

func attrFlowType(val f.Type) otelattr.KeyValue {
	return otelattr.String(attributeKeyFlowType.String(), string(val))
}

func attrSessionID(val uuid.UUID) otelattr.KeyValue {
	return otelattr.String(attributeKeySessionID.String(), val.String())
}

func NewSessionIssued(ctx context.Context, sessionID, identityID uuid.UUID, flow f.Flow) (string, trace.EventOption) {
	var flowType f.Type
	if flow != nil && !reflect.ValueOf(flow).IsNil() {
		flowType = flow.GetType()
	}

	return SessionIssued.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
				attrSessionID(sessionID),
				attrFlowType(flowType),
			)...,
		)
}

func NewRegistrationFailed(ctx context.Context, flow f.Flow) (string, trace.EventOption) {
	var flowType f.Type
	if flow != nil && !reflect.ValueOf(flow).IsNil() {
		flowType = flow.GetType()
	}

	return RegistrationFailed.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				attrFlowType(flowType),
			)...,
		)
}

func NewIdentityCreated(ctx context.Context, identityID uuid.UUID, flow f.Flow) (string, trace.EventOption) {
	var flowType f.Type
	if flow != nil && !reflect.ValueOf(flow).IsNil() {
		flowType = flow.GetType()
	}

	return IdentityCreated.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
				attrFlowType(flowType),
			)...,
		)
}

func NewLoginFailed(ctx context.Context, flow f.Flow) (string, trace.EventOption) {
	var flowType f.Type
	if flow != nil && !reflect.ValueOf(flow).IsNil() {
		flowType = flow.GetType()
	}

	return LoginFailed.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				attrFlowType(flowType),
			)...,
		)
}

func NewSessionRevoked(ctx context.Context, sessionID, identityID uuid.UUID) (string, trace.EventOption) {
	return SessionRevoked.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
				attrSessionID(sessionID),
			)...,
		)
}
