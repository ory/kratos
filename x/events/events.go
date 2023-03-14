package events

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/ory/x/otelx/semconv"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	SignupSuccessful      = "SignupSuccessful"
	SignupFailed          = "SignupFailed"
	LoginSuccessful       = "LoginSuccessful"
	LoginFailed           = "LoginFailed"
	MFAPrompted           = "MFAPrompted"
	MFASuccessful         = "MFASuccessful"
	MFAFailed             = "MFAFailed"
	RecoveryStarted       = "RecoveryStarted"
	RecoverySuccessful    = "RecoverySuccessful"
	VerificationStarted   = "VerificationStarted"
	VerificationSuccesful = "VerificationSuccesful"
	ProfileUpdated        = "ProfileUpdated"
	ProfileUpdateFailed   = "ProfileUpdateFailed"
	CredentialsAdded      = "CredentialsAdded"
	CredentialsRemoved    = "CredentialsRemoved"
	AccountLinked         = "AccountLinked"
	SignOut               = "SignOut"
)

const (
	AdminIdentityCreated        = "AdminIdentityCreated"
	AdminIdentitiesListed       = "AdminIdentitiesListed"
	AdminIdentityViewed         = "AdminIdentityViewed"
	AdminIdentityUpdated        = "AdminIdentityUpdated"
	AdminAccountRecoveryStarted = "AdminAccountRecoveryStarted"
	AdminSessionTerminated      = "AdminSessionTerminated"
)

type NetworkIDProvider interface {
	NetworkID(ctx context.Context) uuid.UUID
}

// Add adds an event to the current span in the context.
func Add(ctx context.Context, n NetworkIDProvider, event string, opt ...trace.EventOption) {
	trace.SpanFromContext(ctx).AddEvent(
		event,
		append(opt, trace.WithAttributes(
			attribute.String(semconv.AttrNID, n.NetworkID(ctx).String()),
		))...,
	)
}
