// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"strings"
	"unicode"

	"github.com/nyaruka/phonenumbers"
	"github.com/pkg/errors"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// CanonicalizeIdentifier applies the Unicode normalization that all
// identifier formats share: NFKC compatibility decomposition followed by
// removal of Unicode format characters (category Cf — zero-width spaces,
// soft hyphens, the BOM, bidi controls, and similar invisibles). Callers
// that compare user-supplied strings for equality (uniqueness checks,
// allowlist matches, identifier lookups) must run this before the
// comparison; otherwise visually identical inputs collide as distinct
// strings.
//
// Run this before lowercasing so the result is stable regardless of the
// casing of compatibility-decomposable characters.
func CanonicalizeIdentifier(value string) string {
	stripInvisible := transform.Chain(norm.NFKC, runes.Remove(runes.In(unicode.Cf)))
	out, _, err := transform.String(stripInvisible, value)
	if err != nil {
		return value
	}
	return out
}

// NormalizeEmailIdentifier normalizes an email address.
func NormalizeEmailIdentifier(value string) string {
	value = CanonicalizeIdentifier(value)
	if strings.Contains(value, "@") {
		value = strings.TrimSpace(strings.ToLower(value))
	}
	return value
}

// NormalizePhoneIdentifier normalizes a phone number.
func NormalizePhoneIdentifier(value string) string {
	value = CanonicalizeIdentifier(value)
	if number, err := phonenumbers.Parse(value, ""); err == nil && phonenumbers.IsValidNumber(number) {
		value = phonenumbers.Format(number, phonenumbers.E164)
	}
	return value
}

// NormalizeOtherIdentifier normalizes an identifier that is not an email or phone number.
func NormalizeOtherIdentifier(value string) string {
	return strings.TrimSpace(strings.ToLower(CanonicalizeIdentifier(value)))
}

// GracefulNormalization normalizes an identifier based on the format.
//
// Supported formats are:
//
// - email
// - phone
// - username
func GracefulNormalization(value string) string {
	value = CanonicalizeIdentifier(value)
	if number, err := phonenumbers.Parse(value, ""); err == nil && phonenumbers.IsValidNumber(number) {
		return phonenumbers.Format(number, phonenumbers.E164)
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
		return NormalizeEmailIdentifier(value), nil
	case "sms":
		number, err := phonenumbers.Parse(CanonicalizeIdentifier(value), "")
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
		return NormalizeOtherIdentifier(value), nil
	}
}
