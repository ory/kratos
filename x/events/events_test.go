// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/kratos/identity"
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
