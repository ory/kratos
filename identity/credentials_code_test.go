// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentialsCodeAddressUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CredentialsCodeAddress
		wantErr bool
	}{
		{
			name:  "valid email address",
			input: `{"channel": "email", "address": "user@example.com"}`,
			want: CredentialsCodeAddress{
				Channel: CodeChannelEmail,
				Address: "user@example.com",
			},
			wantErr: false,
		},
		{
			name:  "valid SMS address",
			input: `{"channel": "sms", "address": "+1234567890"}`,
			want: CredentialsCodeAddress{
				Channel: CodeChannelSMS,
				Address: "+1234567890",
			},
			wantErr: false,
		},
		{
			name:    "invalid address type",
			input:   `{"channel": "invalid", "address": "user@example.com"}`,
			want:    CredentialsCodeAddress{},
			wantErr: true,
		},
		{
			name:    "missing channel field",
			input:   `{"address": "user@example.com"}`,
			want:    CredentialsCodeAddress{},
			wantErr: true,
		},
		{
			name:    "invalid JSON structure",
			input:   `{"channel": "email", "address": "user@example.com"`,
			want:    CredentialsCodeAddress{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got CredentialsCodeAddress
			err := json.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNewCodeAddressType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CodeChannel
		wantErr bool
	}{
		{
			name:    "valid email address type",
			input:   "email",
			want:    CodeChannelEmail,
			wantErr: false,
		},
		{
			name:    "valid SMS address type",
			input:   "sms",
			want:    CodeChannelSMS,
			wantErr: false,
		},
		{
			name:    "invalid address type",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCodeChannel(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
