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
	SessionIssued         semconv.Event = "SessionIssued"
	SessionChanged        semconv.Event = "SessionChanged"
	SessionRevoked        semconv.Event = "SessionRevoked"
	SessionChecked        semconv.Event = "SessionChecked"
	SessionTokenizedAsJWT semconv.Event = "SessionTokenizedAsJWT"
	RegistrationFailed    semconv.Event = "RegistrationFailed"
	RegistrationSucceeded semconv.Event = "RegistrationSucceeded"
	LoginFailed           semconv.Event = "LoginFailed"
	LoginSucceeded        semconv.Event = "LoginSucceeded"
	SettingsFailed        semconv.Event = "SettingsFailed"
	SettingsSucceeded     semconv.Event = "SettingsSucceeded"
	RecoveryFailed        semconv.Event = "RecoveryFailed"
	RecoverySucceeded     semconv.Event = "RecoverySucceeded"
	VerificationFailed    semconv.Event = "VerificationFailed"
	VerificationSucceeded semconv.Event = "VerificationSucceeded"
	IdentityCreated       semconv.Event = "IdentityCreated"
	IdentityUpdated       semconv.Event = "IdentityUpdated"
	WebhookDelivered      semconv.Event = "WebhookDelivered"
	WebhookSucceeded      semconv.Event = "WebhookSucceeded"
	WebhookFailed         semconv.Event = "WebhookFailed"
)

const (
	attributeKeySessionID                       semconv.AttributeKey = "SessionID"
	attributeKeySessionAAL                      semconv.AttributeKey = "SessionAAL"
	attributeKeySelfServiceFlowType             semconv.AttributeKey = "SelfServiceFlowType"
	attributeKeySelfServiceMethodUsed           semconv.AttributeKey = "SelfServiceMethodUsed"
	attributeKeySelfServiceSSOProviderUsed      semconv.AttributeKey = "SelfServiceSSOProviderUsed"
	attributeKeyLoginRequestedAAL               semconv.AttributeKey = "LoginRequestedAAL"
	attributeKeyLoginRequestedPrivilegedSession semconv.AttributeKey = "LoginRequestedPrivilegedSession"
	attributeKeyTokenizedSessionTTL             semconv.AttributeKey = "TokenizedSessionTTL"
	attributeKeyWebhookURL                      semconv.AttributeKey = "WebhookURL"
	attributeKeyWebhookRequestBody              semconv.AttributeKey = "WebhookRequestBody"
	attributeKeyWebhookResponseBody             semconv.AttributeKey = "WebhookResponseBody"
	attributeKeyWebhookResponseStatusCode       semconv.AttributeKey = "WebhookResponseStatusCode"
	attributeKeyWebhookAttemptNumber            semconv.AttributeKey = "WebhookAttemptNumber"
	attributeKeyWebhookRequestID                semconv.AttributeKey = "WebhookRequestID"
)

func attrSessionID(val uuid.UUID) otelattr.KeyValue {
	return otelattr.String(attributeKeySessionID.String(), val.String())
}

func attrTokenizedSessionTTL(ttl time.Duration) otelattr.KeyValue {
	return otelattr.String(attributeKeyTokenizedSessionTTL.String(), ttl.String())
}

func attrSessionAAL(val string) otelattr.KeyValue {
	return otelattr.String(attributeKeySessionAAL.String(), val)
}

func attLoginRequestedAAL(val string) otelattr.KeyValue {
	return otelattr.String(attributeKeyLoginRequestedAAL.String(), val)
}

func attLoginRequestedPrivilegedSession(val bool) otelattr.KeyValue {
	return otelattr.Bool(attributeKeyLoginRequestedPrivilegedSession.String(), val)
}

func attrSelfServiceFlowType(val string) otelattr.KeyValue {
	return otelattr.String(attributeKeySelfServiceFlowType.String(), val)
}

func attrSelfServiceMethodUsed(val string) otelattr.KeyValue {
	return otelattr.String(attributeKeySelfServiceMethodUsed.String(), val)
}

func attrSelfServiceSSOProviderUsed(val string) otelattr.KeyValue {
	return otelattr.String(attributeKeySelfServiceSSOProviderUsed.String(), val)
}

func attrWebhookURL(URL *url.URL) otelattr.KeyValue {
	return otelattr.String(attributeKeyWebhookURL.String(), URL.Redacted())
}

func attrWebhookReq(body []byte) otelattr.KeyValue {
	return otelattr.String(attributeKeyWebhookRequestBody.String(), string(body))
}

func attrWebhookRes(body []byte) otelattr.KeyValue {
	return otelattr.String(attributeKeyWebhookResponseBody.String(), string(body))
}

func attrWebhookStatus(status int) otelattr.KeyValue {
	return otelattr.Int(attributeKeyWebhookResponseStatusCode.String(), status)
}

func attrWebhookAttempt(n int) otelattr.KeyValue {
	return otelattr.Int(attributeKeyWebhookAttemptNumber.String(), n)
}

func attrWebhookRequestID(id uuid.UUID) otelattr.KeyValue {
	return otelattr.String(attributeKeyWebhookRequestID.String(), id.String())
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

func NewWebhookDelivered(ctx context.Context, URL *url.URL, reqBody []byte, status int, resBody []byte, attempt int, requestID uuid.UUID) (string, trace.EventOption) {
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
			)...,
		)
}

func NewWebhookSucceeded(ctx context.Context) (string, trace.EventOption) {
	return WebhookSucceeded.String(),
		trace.WithAttributes(semconv.AttributesFromContext(ctx)...)
}

func NewWebhookFailed(ctx context.Context, err error) (string, trace.EventOption) {
	return WebhookFailed.String(),
		trace.WithAttributes(
			append(
				semconv.AttributesFromContext(ctx),
				otelattr.String("Error", err.Error()),
			)...,
		)
}
