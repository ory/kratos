// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"strings"
	"unicode"

	"golang.org/x/net/idna"

	"github.com/ory/jsonschema/v3"
)

func init() {
	// email_loose accepts non-RFC carrier addresses (e.g. "foo.@docomo.ne.jp")
	// that the strict RFC 5322 "email" format rejects, while still rejecting
	// values that are unsafe in an SMTP envelope or an email header.
	jsonschema.Formats["email_loose"] = func(v interface{}) bool {
		s, ok := v.(string)
		if !ok {
			// Non-string values are handled by the "type" keyword.
			return true
		}
		return IsEmailAddress(s)
	}
}

// emailLocalAtext is the RFC 5322 "atext" punctuation permitted in the local
// part. It deliberately excludes the RFC "specials" — "()<>[]:;@\,\"" — and all
// whitespace/control characters, which could break out of an SMTP envelope or
// inject an email header.
const emailLocalAtext = "!#$%&'*+/=?^_`{|}~.-"

// maxEmailAddressLength caps accepted addresses. 320 is the historical RFC 3696
// bound (64-octet local part + "@" + 255-octet domain).
const maxEmailAddressLength = 320

// IsEmailAddress reports whether s is shaped like an email address using a
// lenient, injection-safe check rather than a strict RFC 5322 parse: a local
// part of letters, digits, and atext punctuation (dot placement is not
// restricted, so carrier addresses like "foo.@docomo.ne.jp" pass), a single
// "@", and a domain that is a valid host name per idna.Lookup. It rejects
// control characters, whitespace, the RFC "specials" that enable SMTP-envelope
// or header injection, and addresses longer than 320 bytes. It is structural
// only and does not verify that the domain resolves or is deliverable.
func IsEmailAddress(s string) bool {
	if len(s) == 0 || len(s) > maxEmailAddressLength {
		return false
	}
	local, domain, ok := strings.Cut(s, "@")
	if !ok || local == "" || domain == "" {
		return false
	}
	for _, r := range local {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
		case strings.ContainsRune(emailLocalAtext, r):
		default:
			return false
		}
	}
	// idna.Lookup rejects any domain that is not a valid host name, so it also
	// gates out embedded "@", control characters, whitespace, and specials.
	if _, err := idna.Lookup.ToASCII(domain); err != nil {
		return false
	}
	return true
}
