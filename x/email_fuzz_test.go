// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"strings"
	"testing"
	"unicode"

	"golang.org/x/net/idna"
)

// FuzzIsEmailAddress asserts the safety invariants of every string that
// IsEmailAddress accepts. These invariants — not the example table — are the
// actual injection-safety contract relied on by the courier SMTP path.
func FuzzIsEmailAddress(f *testing.F) {
	for _, seed := range []string{
		"foo@ory.sh",
		"foo.@docomo.ne.jp",
		".foo@docomo.ne.jp",
		"foo..bar@docomo.ne.jp",
		"a!#$%&'*+/=?^_`{|}~-@ory.sh",
		"föö@münchen.de",
		"foo@ory.sh\r\nBcc: evil@x",
		"foo<bar>@ory.sh",
		"foo@bar@ory.sh",
		"@ory.sh",
		"foo@",
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, s string) {
		if !IsEmailAddress(s) {
			return
		}

		if len(s) > maxEmailAddressLength {
			t.Fatalf("accepted over-long address (%d bytes)", len(s))
		}
		if strings.Count(s, "@") != 1 {
			t.Fatalf("accepted address without exactly one @: %q", s)
		}
		local, domain, _ := strings.Cut(s, "@")
		if local == "" || domain == "" {
			t.Fatalf("accepted empty local part or domain: %q", s)
		}
		for _, r := range s {
			if r < 0x20 || r == 0x7f || unicode.IsSpace(r) || unicode.IsControl(r) {
				t.Fatalf("accepted control/whitespace rune %U in %q", r, s)
			}
			if r != '@' && strings.ContainsRune("()<>[]:;@\\,\"", r) {
				t.Fatalf("accepted RFC specials rune %q in %q", r, s)
			}
		}
		if _, err := idna.Lookup.ToASCII(domain); err != nil {
			t.Fatalf("accepted domain that idna rejects (%v) in %q", err, s)
		}
	})
}
