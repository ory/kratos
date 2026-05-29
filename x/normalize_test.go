// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// defaultIgnorableBypassRunes is the set of code points called out in the
// 2026-05-23 security report as surviving the Cf-only strip. Each rune has
// the Default_Ignorable_Code_Point property but lives in a category outside
// Cf (Lo, Mn, or Cn), so the predicate from PR #11885 misses them and the
// duplicate-account bypass still works.
var defaultIgnorableBypassRunes = []struct {
	name string
	r    rune
}{
	{"Hangul Filler U+3164 (Lo)", 'ㅤ'},
	{"Hangul Choseong Filler U+115F (Lo)", 'ᅟ'},
	{"Hangul Jungseong Filler U+1160 (Lo)", 'ᅠ'},
	{"Khmer Vowel Inherent AQ U+17B4 (Mn)", '឴'},
	{"Khmer Vowel Inherent AA U+17B5 (Mn)", '឵'},
	{"Combining Grapheme Joiner U+034F (Mn)", '͏'},
	{"Reserved U+2065 (Cn, in DI_CP set)", '⁥'},
	{"Variation Selector-16 U+FE0F (Mn)", '️'},
	{"Mongolian Free Variation Selector One U+180B (Mn)", '᠋'},
	{"Variation Selector-17 U+E0100 (Mn)", '\U000E0100'},
}

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
		{"Alice\u00AD@example.com", "alice@example.com"},       // soft hyphen
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
		{" UsERNaME ", "username"},

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

// TestCanonicalizeIdentifier_Concurrent exercises the canonicalization
// path from many goroutines so the race detector catches any shared
// mutable state in the underlying transformer. A package-level
// transform.Chain would race on its internal link buffers; a per-call
// construction must not.
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

// TestNormalize_DefaultIgnorableBypass asserts that every Default_Ignorable_Code_Point
// rune called out in the 2026-05-23 security report is stripped — not just
// the Format (Cf) subset that PR #11885 covered. Each rune is injected mid-
// identifier; the canonical form must equal the rune-free baseline so that
// the UNIQUE constraint on identity_credential_identifiers catches the
// duplicate registration.
func TestNormalize_DefaultIgnorableBypass(t *testing.T) {
	for _, tc := range defaultIgnorableBypassRunes {
		t.Run(tc.name, func(t *testing.T) {
			// Username path.
			gotUser := NormalizeOtherIdentifier("alic" + string(tc.r) + "e")
			assert.Equal(t, "alice", gotUser,
				"username with %s must canonicalize to baseline", tc.name)

			// Email path.
			gotEmail := NormalizeEmailIdentifier("alic" + string(tc.r) + "e@example.com")
			assert.Equal(t, "alice@example.com", gotEmail,
				"email with %s must canonicalize to baseline", tc.name)

			// Phone path: the rune must be removed so phonenumbers.Parse
			// produces the canonical E.164 form.
			gotPhone := NormalizePhoneIdentifier("+1 650-253" + string(tc.r) + "-0000")
			assert.Equal(t, "+16502530000", gotPhone,
				"phone with %s must canonicalize to baseline", tc.name)

			// GracefulNormalization must agree with the format-specific funcs.
			assert.Equal(t, "alice", GracefulNormalization("alic"+string(tc.r)+"e"))
			assert.Equal(t, "alice@example.com",
				GracefulNormalization("alic"+string(tc.r)+"e@example.com"))
			assert.Equal(t, "+16502530000",
				GracefulNormalization("+1 650-253"+string(tc.r)+"-0000"))
		})
	}
}

// TestNormalize_PoisonedRuneStillFolds guards against a duplicate-account
// bypass in the graceful fallback: when an identifier carries a disallowed
// rune AND uppercase or fullwidth characters, stripping the rune is not
// enough — the result must still be lowercased, width-folded, and
// NFKC-composed so it collapses onto the clean baseline. Otherwise an
// attacker registers `ALICE<ZWSP>smith`, it stores as `ALICEsmith`, and the
// UNIQUE constraint never sees it as a duplicate of `alicesmith`.
func TestNormalize_PoisonedRuneStillFolds(t *testing.T) {
	zwsp := string(rune(0x200B))
	tests := []struct {
		name     string
		input    string
		baseline string
	}{
		{"uppercase + zero-width space", "ALICE" + zwsp + "SMITH", "alicesmith"},
		{"fullwidth + zero-width space", "ＡＬＩＣＥ" + zwsp + "smith", "alicesmith"},
		{"uppercase + soft hyphen", "FOO" + string(rune(0x00AD)) + "BAR", "foobar"},
		{"fullwidth digits in email + ZWSP", "ＡＬ" + zwsp + "ICE@example.com", "alice@example.com"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.baseline, NormalizeOtherIdentifier(tc.input))
			assert.Equal(t, NormalizeOtherIdentifier(tc.baseline), NormalizeOtherIdentifier(tc.input))
		})
	}
}

// TestNormalize_EmojiPreserved asserts that emoji and other pictographic
// symbols survive normalization rather than being stripped. They are only
// lowercased and NFKC-composed, so an identifier carrying an emoji stays
// distinct from the emoji-free baseline.
func TestNormalize_EmojiPreserved(t *testing.T) {
	t.Run("emoji is preserved", func(t *testing.T) {
		assert.Equal(t, "alice🎉", NormalizeOtherIdentifier("alice🎉"))
	})

	t.Run("emoji identifier is distinct from baseline", func(t *testing.T) {
		assert.NotEqual(t,
			NormalizeOtherIdentifier("alice"),
			NormalizeOtherIdentifier("alice🎉"),
		)
	})
}

// TestNormalize_RTLScripts asserts that identifiers containing RTL
// script letters (Hebrew, Arabic) — whether pure-RTL or mixed with LTR —
// canonicalize as distinct identifiers. The profile intentionally omits
// the RFC 5893 bidi rule so that international users can register email
// addresses and usernames in their native script.
func TestNormalize_RTLScripts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"pure Hebrew username", "דוד", "דוד"},
		{"pure Arabic username", "محمد", "محمد"},
		{"pure Hebrew email", "דוד@example.com", "דוד@example.com"},
		{"pure Arabic email", "محمد@example.com", "محمد@example.com"},
		{"mixed LTR+Hebrew username", "aliceא", "aliceא"},
		{"mixed LTR+Hebrew email", "alice.דוד@example.com", "alice.דוד@example.com"},
		{"uppercase mixed lowercases", "ALICEא", "aliceא"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, NormalizeOtherIdentifier(tc.input))
		})
	}

	t.Run("mixed-script identifier is distinct from LTR-only baseline", func(t *testing.T) {
		// `aliceא` and `alice` must NOT collide — Hebrew alef is not a
		// homograph for any Latin letter, so they are different users.
		assert.NotEqual(t,
			NormalizeOtherIdentifier("alice"),
			NormalizeOtherIdentifier("aliceא"),
		)
	})

	t.Run("bidi control characters are still stripped", func(t *testing.T) {
		// Even without `precis.BidiRule`, explicit bidi-control
		// characters (LRE, PDF, …) are disallowed at the rune level so
		// they cannot be used to spoof the rendered identifier. The
		// controls are built from code points so the test source itself
		// stays free of bidirectional control characters (gosec G116).
		lre, pdf := string(rune(0x202A)), string(rune(0x202C))
		assert.Equal(t, "alice@example.com",
			NormalizeEmailIdentifier(lre+"alice@example.com"+pdf))
	})
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
