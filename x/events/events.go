// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"

	"github.com/gofrs/uuid"
	otelattr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
	attributeKeySessionID semconv.AttributeKey = "SessionID"
)

func attrSessionID(val uuid.UUID) otelattr.KeyValue {
	return otelattr.String(attributeKeySessionID.String(), val.String())
}

func NewSessionIssued(ctx context.Context, sessionID, identityID uuid.UUID) (string, trace.EventOption) {
	return SessionIssued.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
				attrSessionID(sessionID),
			)...,
		)
}

func NewRegistrationFailed(ctx context.Context) (string, trace.EventOption) {
	return RegistrationFailed.String(),
		trace.WithAttributes(
			semconv.AttributesFromContext(ctx)...,
		)
}

func NewIdentityCreated(ctx context.Context, identityID uuid.UUID) (string, trace.EventOption) {
	return IdentityCreated.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
			)...,
		)
}

func NewLoginFailed(ctx context.Context) (string, trace.EventOption) {
	return LoginFailed.String(),
		trace.WithAttributes(
			semconv.AttributesFromContext(ctx)...,
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
