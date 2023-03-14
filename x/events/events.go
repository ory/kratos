package events

import (
	"context"

	"github.com/ory/kratos/x"
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

// Add adds an event to the current span in the context.
func Add(ctx context.Context, n x.NetworkIDProvider, event string, opt ...attribute.KeyValue) {
	trace.SpanFromContext(ctx).AddEvent(
		event,
		trace.WithAttributes(
			append(opt, attribute.String(semconv.AttrNID, n.NetworkID(ctx).String()))...,
		),
	)
}
