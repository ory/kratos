// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeEmailIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  EXAMPLE@DOMAIN.COM  ", "example@domain.com"},
		{"user@domain.com", "user@domain.com"},
		{"invalid-email", "invalid-email"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, NormalizeEmailIdentifier(test.input))
	}
}

func TestNormalizePhoneIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+1 650-253-0000", "+16502530000"},
		{"+1 (650) 253-0000", "+16502530000"},
		{"invalid-phone", "invalid-phone"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, NormalizePhoneIdentifier(test.input))
	}
}

func TestNormalizeOtherIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  username  ", "username"},
		{"user123", "user123"},
		{"  ", ""},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, NormalizeOtherIdentifier(test.input))
	}
}

func TestGracefulNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+1 650-253-0000", "+16502530000"},
		{"  EXAMPLE@DOMAIN.COM  ", "example@domain.com"},
		{"  username  ", "username"},
		{"invalid-phone", "invalid-phone"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, GracefulNormalization(test.input))
	}
}

func TestNormalizeIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		format   string
		expected string
		err      bool
	}{
		{"  EXAMPLE@DOMAIN.COM  ", "email", "example@domain.com", false},
		{"+1 650-253-0000", "sms", "+16502530000", false},
		{"  username  ", "username", "username", false},
		{"invalid-phone", "sms", "", true},
	}

	for _, test := range tests {
		result, err := NormalizeIdentifier(test.input, test.format)
		if test.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		}
	}
}
