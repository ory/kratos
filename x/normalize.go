// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"fmt"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// NormalizeEmailIdentifier normalizes an email address.
func NormalizeEmailIdentifier(value string) string {
	if strings.Contains(value, "@") {
		value = strings.TrimSpace(strings.ToLower(value))
	}
	return value
}

// NormalizePhoneIdentifier normalizes a phone number.
func NormalizePhoneIdentifier(value string) string {
	if number, err := phonenumbers.Parse(value, ""); err == nil && phonenumbers.IsValidNumber(number) {
		value = phonenumbers.Format(number, phonenumbers.E164)
	}
	return value
}

// NormalizeOtherIdentifier normalizes an identifier that is not an email or phone number.
func NormalizeOtherIdentifier(value string) string {
	return strings.TrimSpace(value)
}

// GracefulNormalization normalizes an identifier based on the format.
//
// Supported formats are:
//
// - email
// - phone
// - username
func GracefulNormalization(value string) string {
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
	case "tel":
		fallthrough
	case "phone":
		fallthrough
	case "sms":
		phoneNumber, err := phonenumbers.Parse(fmt.Sprintf("%s", value), "")
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("+%d%d", *phoneNumber.CountryCode, *phoneNumber.NationalNumber), nil
	case "username":
		fallthrough
	default:
		return NormalizeOtherIdentifier(value), nil
	}
}
