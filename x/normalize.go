// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"net/mail"
	"strings"
	"unicode"

	"github.com/nyaruka/phonenumbers"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/secure/precis"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/unicode/rangetable"
)

// disallowedIdentifierRunes is the explicit set of rune categories the
// identifier profiles reject in addition to the FreeformClass default
// disposition.
var disallowedIdentifierRunes = rangetable.Merge(
	unicode.Bidi_Control,                       // Directional overrides such as RLO/LRO that can spoof display order.
	unicode.Other_Default_Ignorable_Code_Point, // Invisible code points like the zero-width joiner that browsers render as nothing.
	unicode.Other,                              // Control, format, surrogate, private-use, and unassigned code points.
	unicode.Noncharacter_Code_Point,            // Permanently reserved code points (e.g. U+FFFE) that never represent a character.
)

// emailProfile builds on PRECIS FreeformClass — it permits the punctuation
// and spaces that emails need (`@`, `.`, `+`, `-`, `_`) while disallowing
// every Default_Ignorable_Code_Point, bidi control, and pictographic symbol.
//
// `precis.BidiRule` (RFC 5893) is enforced in strict mode so that callers
// like `NormalizeIdentifier("email", …)` reject mixed LTR+RTL inputs at
// registration time.
var emailProfile = precis.NewFreeform(
	precis.FoldWidth,
	precis.LowerCase(),
	precis.DisallowEmpty,
	precis.BidiRule,
	precis.Norm(norm.NFKC),
	precis.Disallow(runes.In(disallowedIdentifierRunes)),
)

// phoneProfile is FreeformClass plus width-fold only. Phone numbers reach
// phonenumbers.Parse which performs its own canonicalization to E.164, so
// the precis step here only width-folds fullwidth digits and removes any
// disallowed runes.
var phoneProfile = precis.NewFreeform(precis.FoldWidth)

func applyPRECIS(profile *precis.Profile, value string, graceful bool) (string, error) {
	value = strings.TrimSpace(value)
	out, err := profile.String(value)
	if err == nil {
		return out, nil
	}
	// Strict mode reports any PRECIS error (disallowed rune or, for the
	// email profile, an RFC 5893 bidi-rule violation) to the caller.
	if !graceful {
		return "", err
	}

	// Graceful mode: strip the runes the profile forbids, then re-run the
	// normalization pipeline on the cleaned value so casing, width folding,
	// and NFKC still apply. Skipping the re-run would leave an uppercase or
	// fullwidth poisoned variant (e.g. "ALICE​smith") distinct from the
	// clean baseline ("alicesmith") and reopen the duplicate-account bypass.
	// The bidi-less companion profile is used so mixed LTR+RTL input also
	// normalizes instead of failing the bidi rule a second time.
	chain := transform.Chain(
		norm.NFKC,
		runes.Remove(runes.Predicate(func(r rune) bool { return !profile.Allowed().Contains(r) })),
		cases.Lower(language.Und),
	)
	stripped, _, err := transform.String(chain, value)
	if err != nil {
		return value, nil
	}
	return stripped, nil
}

// canonicalizeIdentifier gracefully applies the email profile to the string,
// stripping non-printable characters and emojis.
func canonicalizeIdentifier(value string) string {
	v, _ := applyPRECIS(emailProfile, value, true)
	return v
}

// NormalizeEmailIdentifier normalizes an email address.
func NormalizeEmailIdentifier(value string) string {
	return canonicalizeIdentifier(value)
}

// NormalizePhoneIdentifier normalizes a phone number.
func NormalizePhoneIdentifier(value string) string {
	value, _ = applyPRECIS(phoneProfile, value, true)
	if number, err := phonenumbers.Parse(value, ""); err == nil && phonenumbers.IsValidNumber(number) {
		value = phonenumbers.Format(number, phonenumbers.E164)
	}
	return value
}

// NormalizeOtherIdentifier normalizes an identifier that is not an email or phone number.
func NormalizeOtherIdentifier(value string) string {
	return canonicalizeIdentifier(value)
}

// GracefulNormalization normalizes an identifier based on the format.
//
// Supported formats are:
//
// - email
// - phone
// - username
func GracefulNormalization(value string) string {
	value = canonicalizeIdentifier(value)
	if number, err := phonenumbers.Parse(value, ""); err == nil && phonenumbers.IsValidNumber(number) {
		return NormalizePhoneIdentifier(value)
	} else if strings.Contains(value, "@") {
		return NormalizeEmailIdentifier(value)
	}
	return NormalizeOtherIdentifier(value)
}

// NormalizeIdentifier normalizes an identifier based on the format.
//
// Supported formats are:
//
// - email
// - phone
// - username
func NormalizeIdentifier(value, format string) (string, error) {
	switch format {
	case "email":
		value, err := applyPRECIS(emailProfile, value, false)
		if err != nil {
			return "", err
		}
		email, err := mail.ParseAddress(value)
		if err != nil {
			return "", err
		}
		return email.Address, nil

	case "sms":
		value, err := applyPRECIS(phoneProfile, value, false)
		if err != nil {
			return "", err
		}
		number, err := phonenumbers.Parse(value, "")
		if err != nil {
			return "", err
		}
		if !phonenumbers.IsValidNumber(number) {
			return "", errors.New("the provided number is not a valid phone number")
		}
		return phonenumbers.Format(number, phonenumbers.E164), nil

	case "username":
		fallthrough
	default:
		return applyPRECIS(emailProfile, value, true)
	}
}
