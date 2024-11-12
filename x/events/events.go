// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	otelattr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/x/otelx/semconv"
)

const (
	SessionIssued           semconv.Event = "SessionIssued"
	SessionChanged          semconv.Event = "SessionChanged"
	SessionLifespanExtended semconv.Event = "SessionLifespanExtended"
	SessionRevoked          semconv.Event = "SessionRevoked"
	SessionChecked          semconv.Event = "SessionChecked"
	SessionTokenizedAsJWT   semconv.Event = "SessionTokenizedAsJWT"
	RegistrationFailed      semconv.Event = "RegistrationFailed"
	RegistrationSucceeded   semconv.Event = "RegistrationSucceeded"
	LoginFailed             semconv.Event = "LoginFailed"
	LoginSucceeded          semconv.Event = "LoginSucceeded"
	SettingsFailed          semconv.Event = "SettingsFailed"
	SettingsSucceeded       semconv.Event = "SettingsSucceeded"
	RecoveryFailed          semconv.Event = "RecoveryFailed"
	RecoverySucceeded       semconv.Event = "RecoverySucceeded"
	VerificationFailed      semconv.Event = "VerificationFailed"
	VerificationSucceeded   semconv.Event = "VerificationSucceeded"
	IdentityCreated         semconv.Event = "IdentityCreated"
	IdentityUpdated         semconv.Event = "IdentityUpdated"
	IdentityDeleted         semconv.Event = "IdentityDeleted"
	WebhookDelivered        semconv.Event = "WebhookDelivered"
	WebhookSucceeded        semconv.Event = "WebhookSucceeded"
	WebhookFailed           semconv.Event = "WebhookFailed"
)

const (
	AttributeKeySessionID                       semconv.AttributeKey = "SessionID"
	AttributeKeySessionAAL                      semconv.AttributeKey = "SessionAAL"
	AttributeKeySessionExpiresAt                semconv.AttributeKey = "SessionExpiresAt"
	AttributeKeySelfServiceFlowType             semconv.AttributeKey = "SelfServiceFlowType"
	AttributeKeySelfServiceMethodUsed           semconv.AttributeKey = "SelfServiceMethodUsed"
	AttributeKeySelfServiceSSOProviderUsed      semconv.AttributeKey = "SelfServiceSSOProviderUsed"
	AttributeKeyLoginRequestedAAL               semconv.AttributeKey = "LoginRequestedAAL"
	AttributeKeyLoginRequestedPrivilegedSession semconv.AttributeKey = "LoginRequestedPrivilegedSession"
	AttributeKeyTokenizedSessionTTL             semconv.AttributeKey = "TokenizedSessionTTL"
	AttributeKeyWebhookID                       semconv.AttributeKey = "WebhookID"
	AttributeKeyWebhookURL                      semconv.AttributeKey = "WebhookURL"
	AttributeKeyWebhookRequestBody              semconv.AttributeKey = "WebhookRequestBody"
	AttributeKeyWebhookResponseBody             semconv.AttributeKey = "WebhookResponseBody"
	AttributeKeyWebhookResponseStatusCode       semconv.AttributeKey = "WebhookResponseStatusCode"
	AttributeKeyWebhookAttemptNumber            semconv.AttributeKey = "WebhookAttemptNumber"
	AttributeKeyWebhookRequestID                semconv.AttributeKey = "WebhookRequestID"
	AttributeKeyWebhookTriggerID                semconv.AttributeKey = "WebhookTriggerID"
)

func attrSessionID(val uuid.UUID) otelattr.KeyValue {
	return otelattr.String(AttributeKeySessionID.String(), val.String())
}

func attrTokenizedSessionTTL(ttl time.Duration) otelattr.KeyValue {
	return otelattr.String(AttributeKeyTokenizedSessionTTL.String(), ttl.String())
}

func attrSessionAAL(val string) otelattr.KeyValue {
	return otelattr.String(AttributeKeySessionAAL.String(), val)
}

func attLoginRequestedAAL(val string) otelattr.KeyValue {
	return otelattr.String(AttributeKeyLoginRequestedAAL.String(), val)
}

func attSessionExpiresAt(expiresAt time.Time) otelattr.KeyValue {
	return otelattr.String(AttributeKeySessionExpiresAt.String(), expiresAt.String())
}

func attLoginRequestedPrivilegedSession(val bool) otelattr.KeyValue {
	return otelattr.Bool(AttributeKeyLoginRequestedPrivilegedSession.String(), val)
}

func attrSelfServiceFlowType(val string) otelattr.KeyValue {
	return otelattr.String(AttributeKeySelfServiceFlowType.String(), val)
}

func attrSelfServiceMethodUsed(val string) otelattr.KeyValue {
	return otelattr.String(AttributeKeySelfServiceMethodUsed.String(), val)
}

func attrSelfServiceSSOProviderUsed(val string) otelattr.KeyValue {
	return otelattr.String(AttributeKeySelfServiceSSOProviderUsed.String(), val)
}

func attrWebhookID(id string) otelattr.KeyValue {
	return otelattr.String(AttributeKeyWebhookID.String(), id)
}

func attrWebhookURL(URL *url.URL) otelattr.KeyValue {
	return otelattr.String(AttributeKeyWebhookURL.String(), URL.Redacted())
}

func attrWebhookReq(body []byte) otelattr.KeyValue {
	return otelattr.String(AttributeKeyWebhookRequestBody.String(), string(body))
}

func attrWebhookRes(body []byte) otelattr.KeyValue {
	return otelattr.String(AttributeKeyWebhookResponseBody.String(), string(body))
}

func attrWebhookStatus(status int) otelattr.KeyValue {
	return otelattr.Int(AttributeKeyWebhookResponseStatusCode.String(), status)
}

func attrWebhookAttempt(n int) otelattr.KeyValue {
	return otelattr.Int(AttributeKeyWebhookAttemptNumber.String(), n)
}

func attrWebhookRequestID(id uuid.UUID) otelattr.KeyValue {
	return otelattr.String(AttributeKeyWebhookRequestID.String(), id.String())
}

func attrWebhookTriggerID(id uuid.UUID) otelattr.KeyValue {
	return otelattr.String(AttributeKeyWebhookTriggerID.String(), id.String())
}

func NewSessionIssued(ctx context.Context, aal string, sessionID, identityID uuid.UUID) (string, trace.EventOption) {
	return SessionIssued.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
				attrSessionID(sessionID),
				attrSessionAAL(aal),
			)...,
		)
}

func NewSessionChanged(ctx context.Context, aal string, sessionID, identityID uuid.UUID) (string, trace.EventOption) {
	return SessionChanged.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
				attrSessionID(sessionID),
				attrSessionAAL(aal),
			)...,
		)
}

func NewSessionLifespanExtended(ctx context.Context, sessionID, identityID uuid.UUID, newExpiry time.Time) (string, trace.EventOption) {
	return SessionLifespanExtended.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
				attrSessionID(sessionID),
				attSessionExpiresAt(newExpiry),
			)...,
		)
}

type LoginSucceededOpts struct {
	SessionID, IdentityID                       uuid.UUID
	FlowType, RequestedAAL, Method, SSOProvider string
	IsRefresh                                   bool
}

func NewLoginSucceeded(ctx context.Context, o *LoginSucceededOpts) (string, trace.EventOption) {
	return LoginSucceeded.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(o.IdentityID),
				attrSessionID(o.SessionID),
				attrSelfServiceFlowType(o.FlowType),
				attLoginRequestedAAL(o.RequestedAAL),
				attLoginRequestedPrivilegedSession(o.IsRefresh),
				attrSelfServiceMethodUsed(o.Method),
				attrSelfServiceSSOProviderUsed(o.SSOProvider),
			)...,
		)
}

func NewRegistrationSucceeded(ctx context.Context, identityID uuid.UUID, flowType string, method, provider string) (string, trace.EventOption) {
	return RegistrationSucceeded.String(),
		trace.WithAttributes(append(
			semconv.AttributesFromContext(ctx),
			attrSelfServiceFlowType(flowType),
			semconv.AttrIdentityID(identityID),
			attrSelfServiceMethodUsed(method),
			attrSelfServiceSSOProviderUsed(provider),
		)...)
}

func NewRecoverySucceeded(ctx context.Context, identityID uuid.UUID, flowType string, method string) (string, trace.EventOption) {
	return RecoverySucceeded.String(),
		trace.WithAttributes(append(
			semconv.AttributesFromContext(ctx),
			attrSelfServiceFlowType(flowType),
			semconv.AttrIdentityID(identityID),
			attrSelfServiceMethodUsed(method),
		)...)
}

func NewSettingsSucceeded(ctx context.Context, identityID uuid.UUID, flowType string, method string) (string, trace.EventOption) {
	return SettingsSucceeded.String(),
		trace.WithAttributes(append(
			semconv.AttributesFromContext(ctx),
			attrSelfServiceFlowType(flowType),
			semconv.AttrIdentityID(identityID),
			attrSelfServiceMethodUsed(method),
		)...)
}

func NewVerificationSucceeded(ctx context.Context, identityID uuid.UUID, flowType string, method string) (string, trace.EventOption) {
	return VerificationSucceeded.String(),
		trace.WithAttributes(append(
			semconv.AttributesFromContext(ctx),
			attrSelfServiceMethodUsed(method),
			attrSelfServiceFlowType(flowType),
			semconv.AttrIdentityID(identityID),
		)...)
}

func NewRegistrationFailed(ctx context.Context, flowType string, method string) (string, trace.EventOption) {
	return RegistrationFailed.String(),
		trace.WithAttributes(append(
			semconv.AttributesFromContext(ctx),
			attrSelfServiceFlowType(flowType),
			attrSelfServiceMethodUsed(method),
		)...)
}

func NewRecoveryFailed(ctx context.Context, flowType string, method string) (string, trace.EventOption) {
	return RecoveryFailed.String(),
		trace.WithAttributes(append(
			semconv.AttributesFromContext(ctx),
			attrSelfServiceFlowType(flowType),
			attrSelfServiceMethodUsed(method),
		)...)
}

func NewSettingsFailed(ctx context.Context, flowType string, method string) (string, trace.EventOption) {
	return SettingsFailed.String(),
		trace.WithAttributes(append(
			semconv.AttributesFromContext(ctx),
			attrSelfServiceFlowType(flowType),
			attrSelfServiceMethodUsed(method),
		)...)
}

func NewVerificationFailed(ctx context.Context, flowType string, method string) (string, trace.EventOption) {
	return VerificationFailed.String(),
		trace.WithAttributes(append(
			semconv.AttributesFromContext(ctx),
			attrSelfServiceFlowType(flowType),
			attrSelfServiceMethodUsed(method),
		)...)
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

func NewIdentityDeleted(ctx context.Context, identityID uuid.UUID) (string, trace.EventOption) {
	return IdentityDeleted.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
			)...,
		)
}

func NewIdentityUpdated(ctx context.Context, identityID uuid.UUID) (string, trace.EventOption) {
	return IdentityUpdated.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
			)...,
		)
}

func NewLoginFailed(ctx context.Context, flowType string, requestedAAL string, isRefresh bool) (string, trace.EventOption) {
	return LoginFailed.String(),
		trace.WithAttributes(append(
			semconv.AttributesFromContext(ctx),
			attrSelfServiceFlowType(flowType),
			attLoginRequestedAAL(requestedAAL),
			attLoginRequestedPrivilegedSession(isRefresh),
		)...)
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

func NewSessionChecked(ctx context.Context, sessionID, identityID uuid.UUID) (string, trace.EventOption) {
	return SessionChecked.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
				attrSessionID(sessionID),
			)...,
		)
}

func NewSessionJWTIssued(ctx context.Context, sessionID, identityID uuid.UUID, ttl time.Duration) (string, trace.EventOption) {
	return SessionTokenizedAsJWT.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				semconv.AttrIdentityID(identityID),
				attrSessionID(sessionID),
				attrTokenizedSessionTTL(ttl),
			)...,
		)
}

func NewWebhookDelivered(ctx context.Context, URL *url.URL, reqBody []byte, status int, resBody []byte, attempt int, requestID, triggerID uuid.UUID, webhookID string) (string, trace.EventOption) {
	return WebhookDelivered.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				attrWebhookReq(reqBody),
				attrWebhookRes(resBody),
				attrWebhookStatus(status),
				attrWebhookURL(URL),
				attrWebhookAttempt(attempt),
				attrWebhookRequestID(requestID),
				attrWebhookID(webhookID),
				attrWebhookTriggerID(triggerID),
			)...,
		)
}

func NewWebhookSucceeded(ctx context.Context, triggerID uuid.UUID, webhookID string) (string, trace.EventOption) {
	return WebhookSucceeded.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				attrWebhookID(webhookID),
				attrWebhookTriggerID(triggerID),
			)...)
}

func NewWebhookFailed(ctx context.Context, err error, triggerID uuid.UUID, id string) (string, trace.EventOption) {
	return WebhookFailed.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				attrWebhookID(id),
				attrWebhookTriggerID(triggerID),
				otelattr.String("Error", err.Error()),
			)...,
		)
}
