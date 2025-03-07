// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/identity"
)

func TestFindAllIdentifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    *identity.Identity
		expected []Address
	}{
		{
			name: "valid verifiable addresses",
			input: &identity.Identity{
				VerifiableAddresses: []identity.VerifiableAddress{
					{Via: "email", Value: "user@example.com"},
					{Via: "sms", Value: "+1234567890"},
				},
			},
			expected: []Address{
				{Via: identity.CodeChannel("email"), To: "user@example.com"},
				{Via: identity.CodeChannel("sms"), To: "+1234567890"},
			},
		},
		{
			name: "empty verifiable addresses",
			input: &identity.Identity{
				VerifiableAddresses: []identity.VerifiableAddress{},
			},
		},
		{
			name: "verifiable address with empty fields",
			input: &identity.Identity{
				VerifiableAddresses: []identity.VerifiableAddress{
					{Via: "", Value: ""},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindAllIdentifiers(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindCodeAddressCandidates(t *testing.T) {
	tests := []struct {
		name            string
		input           *identity.Identity
		fallbackEnabled bool
		expected        []Address
		found           bool
		wantErr         bool
	}{
		{
			name: "valid credentials with addresses",
			input: &identity.Identity{
				Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypeCodeAuth: {
						Config: []byte(`{"addresses":[{"channel":"email","address":"user@example.com"},{"channel":"sms","address":"+1234567890"}]}`),
					},
				},
			},
			fallbackEnabled: false,
			expected: []Address{
				{Via: identity.CodeChannel("email"), To: "user@example.com"},
				{Via: identity.CodeChannel("sms"), To: "+1234567890"},
			},
			found:   true,
			wantErr: false,
		},
		{
			name: "no credentials, fallback enabled",
			input: &identity.Identity{
				VerifiableAddresses: []identity.VerifiableAddress{
					{Via: "email", Value: "user@example.com"},
					{Via: "sms", Value: "+1234567890"},
				},
			},
			fallbackEnabled: true,
			expected: []Address{
				{Via: identity.CodeChannel("email"), To: "user@example.com"},
				{Via: identity.CodeChannel("sms"), To: "+1234567890"},
			},
			found:   true,
			wantErr: false,
		},
		{
			name: "no credentials, fallback disabled",
			input: &identity.Identity{
				VerifiableAddresses: []identity.VerifiableAddress{
					{Via: "email", Value: "user@example.com"},
					{Via: "sms", Value: "+1234567890"},
				},
			},
			fallbackEnabled: false,
			expected:        nil,
			found:           false,
			wantErr:         false,
		},
		{
			name: "invalid credentials config",
			input: &identity.Identity{
				Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypeCodeAuth: {
						Config: []byte(`invalid`),
					},
				},
			},
			fallbackEnabled: false,
			expected:        nil,
			found:           false,
			wantErr:         true,
		},
		{
			name: "invalid credentials config, fallback enabled, verifiable addresses exist",
			input: &identity.Identity{
				Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypeCodeAuth: {
						Config: []byte(`invalid`),
					},
				},
				VerifiableAddresses: []identity.VerifiableAddress{
					{Via: "email", Value: "user@example.com"},
					{Via: "sms", Value: "+1234567890"},
				},
			},
			fallbackEnabled: true,
			expected: []Address{
				{Via: identity.CodeChannel("email"), To: "user@example.com"},
				{Via: identity.CodeChannel("sms"), To: "+1234567890"},
			},
			found:   true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found, err := FindCodeAddressCandidates(tt.input, tt.fallbackEnabled)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
				assert.Equal(t, tt.found, found)
			}
		})
	}
}
