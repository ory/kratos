// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/x"
	"github.com/ory/x/otelx/semconv"
)

const (
	SignupSuccessful      semconv.Event = semconv.EventIdentityCreated
	SignupFailed          semconv.Event = "SignupFailed"
	LoginSuccessful       semconv.Event = semconv.EventSessionIssued
	LoginFailed           semconv.Event = "LoginFailed"
	MFAPrompted           semconv.Event = "MFAPrompted"
	MFASuccessful         semconv.Event = "MFASuccessful"
	MFAFailed             semconv.Event = "MFAFailed"
	RecoveryStarted       semconv.Event = "RecoveryStarted"
	RecoverySuccessful    semconv.Event = "RecoverySuccessful"
	VerificationStarted   semconv.Event = "VerificationStarted"
	VerificationSuccesful semconv.Event = "VerificationSuccesful"
	ProfileUpdated        semconv.Event = "ProfileUpdated"
	ProfileUpdateFailed   semconv.Event = "ProfileUpdateFailed"
	CredentialsAdded      semconv.Event = "CredentialsAdded"
	CredentialsRemoved    semconv.Event = "CredentialsRemoved"
	AccountLinked         semconv.Event = "AccountLinked"
	SignOut               semconv.Event = "SignOut"
)

const (
	AdminIdentityCreated        = "AdminIdentityCreated"
	AdminIdentitiesListed       = "AdminIdentitiesListed"
	AdminIdentityViewed         = "AdminIdentityViewed"
	AdminIdentityUpdated        = "AdminIdentityUpdated"
	AdminAccountRecoveryStarted = "AdminAccountRecoveryStarted"
	AdminSessionTerminated      = "AdminSessionTerminated"
)

// Add adds an event to the current span in the context.
func Add(ctx context.Context, n x.NetworkIDProvider, event semconv.Event, opt ...attribute.KeyValue) {
	trace.SpanFromContext(semconv.ContextWithAttributes(ctx)).AddEvent(
		string(event),
		trace.WithAttributes(
			opt...,
		),
	)
}
