// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmailAddress(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		want bool
	}{
		// Accepted: normal and non-RFC-but-safe carrier addresses.
		{name: "plain", in: "foo@ory.sh", want: true},
		{name: "subdomain and plus", in: "foo.bar+tag@mail.ory.sh", want: true},
		{name: "docomo trailing dot", in: "foo.@docomo.ne.jp", want: true},
		{name: "docomo double dot", in: "foo..bar@docomo.ne.jp", want: true},
		{name: "docomo leading dot", in: ".foo@docomo.ne.jp", want: true},
		{name: "internationalized local part", in: "föö@ory.sh", want: true},
		{name: "atext punctuation", in: "a!#$%&'*+/=?^_`{|}~-@ory.sh", want: true},
		{name: "single label domain", in: "foo@localhost", want: true},
		{name: "idn domain accepted via idna", in: "foo@münchen.de", want: true},
		{name: "exactly 320 bytes", in: strings.Repeat("a", 313) + "@ory.sh", want: true},
		{name: "321 bytes over cap", in: strings.Repeat("a", 314) + "@ory.sh", want: false},

		// Rejected: not an address.
		{name: "empty", in: "", want: false},
		{name: "no at", in: "invalid-email", want: false},
		{name: "empty local", in: "@ory.sh", want: false},
		{name: "empty domain", in: "foo@", want: false},
		{name: "two at signs", in: "foo@bar@ory.sh", want: false},

		// Rejected: injection / SMTP-envelope specials and whitespace.
		{name: "CR injection", in: "foo@ory.sh\r\nBcc: evil@x", want: false},
		{name: "LF injection", in: "foo@ory.sh\nBcc: evil@x", want: false},
		{name: "space", in: "foo bar@ory.sh", want: false},
		{name: "tab", in: "foo\t@ory.sh", want: false},
		{name: "angle brackets", in: "foo<bar>@ory.sh", want: false},
		{name: "comma", in: "foo@ory.sh,evil@x", want: false},
		{name: "semicolon", in: "foo;bar@ory.sh", want: false},
		{name: "colon", in: "foo:bar@ory.sh", want: false},
		{name: "double quote", in: "\"foo\"@ory.sh", want: false},
		{name: "backslash", in: "foo\\bar@ory.sh", want: false},
		{name: "parenthesis", in: "foo(comment)@ory.sh", want: false},
		{name: "square bracket ip literal", in: "foo@[192.168.0.1]", want: false},

		// Rejected: domain is not a valid host name (idna Lookup profile).
		{name: "bang in domain", in: "foo@bar!baz", want: false},
		{name: "percent in domain", in: "foo@bar%baz", want: false},
		{name: "underscore in domain", in: "foo@bar_baz.com", want: false},
		{name: "backtick in domain", in: "foo@ba`r.com", want: false},
		{name: "over length cap", in: strings.Repeat("a", 310) + "@" + strings.Repeat("b", 20) + ".jp", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, IsEmailAddress(tc.in))
		})
	}
}
