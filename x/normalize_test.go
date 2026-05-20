// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"sync"
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

		// Invisible Unicode characters (category Cf) must not produce a
		// distinct identifier — otherwise an attacker can register a second
		// account with a visually identical email and bypass the duplicate-
		// account check.
		{"alice\u200B@example.com", "alice@example.com"},       // zero-width space
		{"alice\u200C@example.com", "alice@example.com"},       // zero-width non-joiner
		{"alice\u200D@example.com", "alice@example.com"},       // zero-width joiner
		{"alice\u00AD@example.com", "alice@example.com"},       // soft hyphen
		{"alice\uFEFF@example.com", "alice@example.com"},       // zero-width no-break space (BOM)
		{"alice\u2060@example.com", "alice@example.com"},       // word joiner
		{"alice\u180E@example.com", "alice@example.com"},       // Mongolian vowel separator
		{"\u202Aalice@example.com\u202C", "alice@example.com"}, // bidi LRE/PDF
		{"al\u200Bice\u200C@ex\u200Dample.com", "alice@example.com"},

		// NFKC compatibility decomposition collapses fullwidth and ligature
		// lookalikes so they hit the same identifier as the ASCII canonical
		// form.
		{"ａｌｉｃｅ@example.com", "alice@example.com"}, // fullwidth "alice"
		{"oﬃce@example.com", "office@example.com"}, // ligature ffi
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, NormalizeEmailIdentifier(test.input), "input=%q", test.input)
	}
}

func TestNormalizePhoneIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+1 650-253-0000", "+16502530000"},
		{"+1 (650) 253-0000", "+16502530000"},
		{"+44 20 7946 0958", "+442079460958"},
		{"+49 176 671 11 638", "+4917667111638"},
		{"+600136866982", "+60136866982"},
		{"+256 730 691 099", "+256730691099"},
		{"+256730691099", "+256730691099"},
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

		// Invisible characters must not allow a duplicate username either.
		{"al\u200Bice", "alice"}, // U+200B
		{"al\u200Cice", "alice"}, // U+200C
		{"al\u200Dice", "alice"}, // U+200D
		{"al\u00ADice", "alice"}, // U+00AD soft hyphen
		{"al\uFEFFice", "alice"}, // U+FEFF BOM
		{"al\u2060ice", "alice"}, // U+2060 word joiner
		{"al\u180Eice", "alice"}, // U+180E
		{"ａｌｉｃｅ", "alice"},       // fullwidth Latin (NFKC)
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, NormalizeOtherIdentifier(test.input), "input=%q", test.input)
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

// TestCanonicalizeIdentifier_Concurrent exercises CanonicalizeIdentifier from
// many goroutines so the race detector catches any shared mutable state in the
// underlying transformer. A package-level transform.Chain would race on its
// internal link buffers; a per-call construction must not.
func TestCanonicalizeIdentifier_Concurrent(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		input, want string
	}{
		{"alice\u200B@example.com", "alice@example.com"},
		{"al\u200Cice\u200D@ex\u00ADample.com", "alice@example.com"},
		{"ａｌｉｃｅ@example.com", "alice@example.com"},
		{"oﬃce@example.com", "office@example.com"},
		{"al\uFEFFice\u2060@example.com", "alice@example.com"},
	}

	const goroutines = 64
	const iterations = 200

	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range iterations {
				for _, tc := range inputs {
					got := NormalizeEmailIdentifier(tc.input)
					if got != tc.want {
						assert.Equal(t, tc.want, got, "input=%q", tc.input)
						return
					}
				}
			}
		})
	}
	wg.Wait()
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
		{"+256 730 691 099", "sms", "+256730691099", false},
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
