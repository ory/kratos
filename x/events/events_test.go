// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/events"
)

func TestNewJsonnetMappingFailed(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		jsonnetInput  []byte
		jsonnetOutput string
		provider      string
		method        identity.CredentialsType
		expectedAttrs []attribute.KeyValue
	}{
		{
			name:          "With all attributes",
			err:           errors.New("test error"),
			jsonnetInput:  []byte(`{"key": "PII value"}`),
			jsonnetOutput: `{"key": 123}`,
			provider:      "test-provider",
			method:        identity.CredentialsTypeOIDC,
			expectedAttrs: []attribute.KeyValue{
				attribute.String("SelfServiceSSOProviderUsed", "test-provider"),
				attribute.String("SelfServiceMethodUsed", "oidc"),
				attribute.String("ErrorReason", "test error"),
				attribute.String("JsonnetInput", `{
  "key": "string"
}`),
				attribute.String("JsonnetOutput", `{
  "key": "number"
}`),
			},
		},
		{
			name:          "Without JsonnetOutput",
			err:           errors.New("another error"),
			jsonnetInput:  []byte(`{"key": "PII value"}`),
			jsonnetOutput: "",
			provider:      "another-provider",
			method:        identity.CredentialsTypeSAML,
			expectedAttrs: []attribute.KeyValue{
				attribute.String("SelfServiceSSOProviderUsed", "another-provider"),
				attribute.String("SelfServiceMethodUsed", "saml"),
				attribute.String("ErrorReason", "another error"),
				attribute.String("JsonnetInput", `{
  "key": "string"
}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			eventName, opts := events.NewJsonnetMappingFailed(ctx, tt.err, tt.jsonnetInput, tt.jsonnetOutput, tt.provider, tt.method.String())

			assert.Equal(t, events.JsonnetMappingFailed.String(), eventName)

			eventConfig := trace.NewEventConfig(opts)
			assert.ElementsMatch(t, tt.expectedAttrs, eventConfig.Attributes())
		})
	}
}

func TestNewLoginFailed(t *testing.T) {
	ctx := t.Context()
	flowID := uuid.Must(uuid.NewV4())
	identityID := uuid.Must(uuid.NewV4())
	baseErr := errors.New("login failed")

	t.Run("case=without identity ID", func(t *testing.T) {
		eventName, opts := events.NewLoginFailed(ctx, flowID, "browser", "aal1", false, baseErr)

		assert.Equal(t, events.LoginFailed.String(), eventName)

		eventConfig := trace.NewEventConfig(opts)
		attrs := eventConfig.Attributes()

		// Should not contain IdentityID attribute
		for _, attr := range attrs {
			assert.NotEqual(t, "IdentityID", string(attr.Key))
		}

		// Should contain other attributes
		assert.Contains(t, attrs, attribute.String("SelfServiceFlowType", "browser"))
		assert.Contains(t, attrs, attribute.String("LoginRequestedAAL", "aal1"))
		assert.Contains(t, attrs, attribute.Bool("LoginRequestedPrivilegedSession", false))
		assert.Contains(t, attrs, attribute.String("ErrorReason", "login failed"))
	})

	t.Run("case=with identity ID", func(t *testing.T) {
		wrappedErr := x.WrapWithIdentityIDError(baseErr, identityID)
		eventName, opts := events.NewLoginFailed(ctx, flowID, "browser", "aal1", false, wrappedErr)

		assert.Equal(t, events.LoginFailed.String(), eventName)

		eventConfig := trace.NewEventConfig(opts)
		attrs := eventConfig.Attributes()

		assert.Contains(t, attrs, attribute.String("IdentityID", identityID.String()))
		assert.Contains(t, attrs, attribute.String("SelfServiceFlowType", "browser"))
		assert.Contains(t, attrs, attribute.String("ErrorReason", "login failed"))
	})
}

func TestNewRecoveryFailed(t *testing.T) {
	ctx := t.Context()
	flowID := uuid.Must(uuid.NewV4())
	identityID := uuid.Must(uuid.NewV4())
	baseErr := errors.New("recovery failed")

	t.Run("case=without identity ID", func(t *testing.T) {
		eventName, opts := events.NewRecoveryFailed(ctx, flowID, "browser", "code", baseErr)

		assert.Equal(t, events.RecoveryFailed.String(), eventName)

		eventConfig := trace.NewEventConfig(opts)
		attrs := eventConfig.Attributes()

		for _, attr := range attrs {
			assert.NotEqual(t, "IdentityID", string(attr.Key))
		}

		assert.Contains(t, attrs, attribute.String("SelfServiceFlowType", "browser"))
		assert.Contains(t, attrs, attribute.String("SelfServiceMethodUsed", "code"))
	})

	t.Run("case=with identity ID", func(t *testing.T) {
		wrappedErr := x.WrapWithIdentityIDError(baseErr, identityID)
		eventName, opts := events.NewRecoveryFailed(ctx, flowID, "browser", "link", wrappedErr)

		assert.Equal(t, events.RecoveryFailed.String(), eventName)

		eventConfig := trace.NewEventConfig(opts)
		attrs := eventConfig.Attributes()

		assert.Contains(t, attrs, attribute.String("IdentityID", identityID.String()))
		assert.Contains(t, attrs, attribute.String("SelfServiceFlowType", "browser"))
		assert.Contains(t, attrs, attribute.String("SelfServiceMethodUsed", "link"))
	})
}

func TestNewSettingsFailed(t *testing.T) {
	ctx := t.Context()
	flowID := uuid.Must(uuid.NewV4())
	identityID := uuid.Must(uuid.NewV4())
	baseErr := errors.New("settings failed")

	t.Run("case=without identity ID", func(t *testing.T) {
		eventName, opts := events.NewSettingsFailed(ctx, flowID, "browser", "profile", baseErr)

		assert.Equal(t, events.SettingsFailed.String(), eventName)

		eventConfig := trace.NewEventConfig(opts)
		attrs := eventConfig.Attributes()

		for _, attr := range attrs {
			assert.NotEqual(t, "IdentityID", string(attr.Key))
		}

		assert.Contains(t, attrs, attribute.String("SelfServiceFlowType", "browser"))
		assert.Contains(t, attrs, attribute.String("SelfServiceMethodUsed", "profile"))
	})

	t.Run("case=with identity ID", func(t *testing.T) {
		wrappedErr := x.WrapWithIdentityIDError(baseErr, identityID)
		eventName, opts := events.NewSettingsFailed(ctx, flowID, "browser", "password", wrappedErr)

		assert.Equal(t, events.SettingsFailed.String(), eventName)

		eventConfig := trace.NewEventConfig(opts)
		attrs := eventConfig.Attributes()

		assert.Contains(t, attrs, attribute.String("IdentityID", identityID.String()))
		assert.Contains(t, attrs, attribute.String("SelfServiceFlowType", "browser"))
		assert.Contains(t, attrs, attribute.String("SelfServiceMethodUsed", "password"))
	})
}

func TestNewVerificationFailed(t *testing.T) {
	ctx := t.Context()
	flowID := uuid.Must(uuid.NewV4())
	identityID := uuid.Must(uuid.NewV4())
	baseErr := errors.New("verification failed")

	t.Run("case=without identity ID", func(t *testing.T) {
		eventName, opts := events.NewVerificationFailed(ctx, flowID, "browser", "code", baseErr)

		assert.Equal(t, events.VerificationFailed.String(), eventName)

		eventConfig := trace.NewEventConfig(opts)
		attrs := eventConfig.Attributes()

		for _, attr := range attrs {
			assert.NotEqual(t, "IdentityID", string(attr.Key))
		}

		assert.Contains(t, attrs, attribute.String("SelfServiceFlowType", "browser"))
		assert.Contains(t, attrs, attribute.String("SelfServiceMethodUsed", "code"))
	})

	t.Run("case=with identity ID", func(t *testing.T) {
		wrappedErr := x.WrapWithIdentityIDError(baseErr, identityID)
		eventName, opts := events.NewVerificationFailed(ctx, flowID, "browser", "link", wrappedErr)

		assert.Equal(t, events.VerificationFailed.String(), eventName)

		eventConfig := trace.NewEventConfig(opts)
		attrs := eventConfig.Attributes()

		assert.Contains(t, attrs, attribute.String("IdentityID", identityID.String()))
		assert.Contains(t, attrs, attribute.String("SelfServiceFlowType", "browser"))
		assert.Contains(t, attrs, attribute.String("SelfServiceMethodUsed", "link"))
	})
}
