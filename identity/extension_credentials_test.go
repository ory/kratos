// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/jsonschema/v3"
	_ "github.com/ory/jsonschema/v3/fileloader"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/x/snapshotx"
	"github.com/ory/x/sqlxx"
)

func TestSchemaExtensionCredentials(t *testing.T) {
	for k, tc := range []struct {
		expectErr           error
		schema              string
		doc                 string
		expectedIdentifiers []string
		existing            *identity.Credentials
		ct                  identity.CredentialsType
	}{
		{
			doc:                 `{"email":"foo@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			ct:                  identity.CredentialsTypePassword,
		},
		{
			doc:                 `{"emails":["foo😀@ory.sh","foo😀@ory.sh","bar@ory.sh"], "username": "Foo🥳🖐baR"}`,
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"foo😀@ory.sh", "bar@ory.sh", "foo🥳🖐bar"},
			ct:                  identity.CredentialsTypePassword,
		},
		{
			doc:                 `{"emails":["föo@ory.sh","föo@ory.sh","bar@ory.sh"], "username": "foobar"}`,
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"föo@ory.sh", "bar@ory.sh"},
			ct:                  identity.CredentialsTypeWebAuthn,
		},
		{
			doc:                 `{"emails":["FÖÖ@ory.sh","bar@ory.sh"], "username": "foobär"}`,
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"föö@ory.sh", "bar@ory.sh", "foobär"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh"},
			},
			ct: identity.CredentialsTypePassword,
		},
		{
			doc:                 `{"email":"foo@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/webauthn.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			ct:                  identity.CredentialsTypeWebAuthn,
		},
		{
			doc:                 `{"email":"FOO@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/webauthn.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh"},
			},
			ct: identity.CredentialsTypeWebAuthn,
		},
		{
			doc:                 `{"email":"foo@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/code.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			ct:                  identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/code.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh"},
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/code.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh", "foo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"not-foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh","phone":"+49 176 671 11 638"}`,
			schema:              "file://./stub/extension/credentials/code-phone-email.schema.json",
			expectedIdentifiers: []string{"+4917667111638", "foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh", "foo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"not-foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh","phone":"+49 176 671 11 638"}`,
			schema:              "file://./stub/extension/credentials/code-phone-email.schema.json",
			expectedIdentifiers: []string{"+4917667111638", "foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh", "foo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"not-foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh","email2":"FOO@ory.sh","phone":"+49 176 671 11 638"}`,
			schema:              "file://./stub/extension/credentials/code-phone-email.schema.json",
			expectedIdentifiers: []string{"+4917667111638", "foo@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh", "fOo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"not-foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"FOO@ory.sh","email2":"FOO@ory.sh","email3":"bar@ory.sh","phone":"+49 176 671 11 638"}`,
			schema:              "file://./stub/extension/credentials/code-phone-email.schema.json",
			expectedIdentifiers: []string{"+4917667111638", "foo@ory.sh", "bar@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"not-foo@ory.sh", "fOo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"not-foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
		},
		{
			doc:                 `{"email":"bar@ory.sh"}`,
			schema:              "file://./stub/extension/credentials/email.schema.json",
			expectedIdentifiers: []string{"bar@ory.sh"},
			existing: &identity.Credentials{
				Identifiers: []string{"foo@ory.sh"},
				Config:      sqlxx.JSONRawMessage(`{"addresses":[{"channel":"email","address":"foo@ory.sh"}]}`),
			},
			ct: identity.CredentialsTypeCodeAuth,
		},

		{
			doc:                 `{"email":"foo@ory.sh","email2":"foo@ory.sh","email3":"bar@ory.sh","phone":"+49 176 671 11 638"}`,
			schema:              "file://./stub/extension/credentials/code-phone-email.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh", "bar@ory.sh", "+4917667111638"},
			ct:                  identity.CredentialsTypePassword,
		},
		{
			doc:                 `{"email":"foo@ory.sh","phone":"+49 176 671 11 638"}`,
			schema:              "file://./stub/extension/credentials/code-phone-email.schema.json",
			expectedIdentifiers: []string{"foo@ory.sh", "+4917667111638"},
			existing: &identity.Credentials{
				Identifiers: []string{"foo@ory.sh", "+49 176 671 11 638"},
			},
			ct: identity.CredentialsTypePassword,
		},
		{
			doc:       `{"emails":["bar@o😶ry.sh"], "username": "Foo🥳🖐baR"}`,
			schema:    "file://./stub/extension/credentials/multi.schema.json",
			ct:        identity.CredentialsTypePassword,
			expectErr: errors.New("I[#/emails/0] S[#/properties/emails/items/format] \"bar@o😶ry.sh\" is not valid \"email\""),
		},
		{
			doc:       `{"email":"foo@öry.sh"}`,
			schema:    "file://./stub/extension/credentials/webauthn.schema.json",
			expectErr: errors.New(`I[#/email] S[#/properties/email/format] "foo@öry.sh" is not valid "email"`),
			ct:        identity.CredentialsTypeWebAuthn,
		},

		// Invisible and bidi-control characters must be stripped from both
		// emails and usernames. Each injected rune below has been used in
		// real homograph / duplicate-account bypasses: the poisoned input
		// must canonicalize to the same baseline an attacker is trying to
		// impersonate, so the UNIQUE constraint on
		// identity_credential_identifiers rejects the duplicate.
		//
		// The runes are built with fmt.Sprintf("%c", …) on purpose: literal
		// invisible characters in test source are unreviewable, and a literal
		// U+FEFF (BOM) is rejected by the Go compiler outright.

		// Zero-width space (U+200B) inside an email local part.
		{
			doc:                 fmt.Sprintf(`{"emails":["ad%cmin@ory.sh"],"username":"zwsp-email"}`, 0x200B),
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"admin@ory.sh", "zwsp-email"},
			ct:                  identity.CredentialsTypePassword,
		},
		// Several different invisibles (ZWSP, ZWJ, soft hyphen) in one email.
		{
			doc:                 fmt.Sprintf(`{"emails":["a%cd%cmi%cn@ory.sh"],"username":"multi-email"}`, 0x200B, 0x200D, 0x00AD),
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"admin@ory.sh", "multi-email"},
			ct:                  identity.CredentialsTypePassword,
		},
		// BOM (U+FEFF) prefix and word joiner (U+2060) suffix in an email.
		{
			doc:                 fmt.Sprintf(`{"emails":["%cadmin%c@ory.sh"],"username":"bom-email"}`, 0xFEFF, 0x2060),
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"admin@ory.sh", "bom-email"},
			ct:                  identity.CredentialsTypePassword,
		},
		// Right-to-left override (U+202E, Trojan-source style) in an email.
		{
			doc:                 fmt.Sprintf(`{"emails":["ad%cmin@ory.sh"],"username":"rlo-email"}`, 0x202E),
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"admin@ory.sh", "rlo-email"},
			ct:                  identity.CredentialsTypePassword,
		},
		// Zero-width space (U+200B) inside a username.
		{
			doc:                 fmt.Sprintf(`{"emails":["zwsp-user@ory.sh"],"username":"ad%cmin"}`, 0x200B),
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"zwsp-user@ory.sh", "admin"},
			ct:                  identity.CredentialsTypePassword,
		},
		// Default-ignorable runes outside category Cf in a username: ZWNJ,
		// combining grapheme joiner (U+034F), Hangul filler (U+3164), and
		// variation selector-16 (U+FE0F) all collapse to the baseline.
		{
			doc:                 fmt.Sprintf(`{"emails":["multi-user@ory.sh"],"username":"a%cd%cm%ci%cn"}`, 0x200C, 0x034F, 0x3164, 0xFE0F),
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"multi-user@ory.sh", "admin"},
			ct:                  identity.CredentialsTypePassword,
		},
		// Bidi override (U+202E) plus word joiner (U+2060) and soft hyphen
		// (U+00AD) wrapping a username.
		{
			doc:                 fmt.Sprintf(`{"emails":["rlo-user@ory.sh"],"username":"%cad%cmin%c"}`, 0x202E, 0x2060, 0x00AD),
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"rlo-user@ory.sh", "admin"},
			ct:                  identity.CredentialsTypePassword,
		},
		// Two differently-poisoned variants of the same email must dedupe to
		// a single identifier — this is the duplicate-account bypass itself.
		{
			doc:                 fmt.Sprintf(`{"emails":["ad%cmin@ory.sh","ad%cmin@ory.sh"],"username":"dedup"}`, 0x200B, 0x00AD),
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"admin@ory.sh", "dedup"},
			ct:                  identity.CredentialsTypePassword,
		},
		// Boundary: NFKC does not fold script confusables. A Cyrillic "а"
		// (U+0430) is a distinct letter from Latin "a", so this username
		// stays distinct. Confusable-script detection is intentionally out
		// of scope for identifier normalization.
		{
			doc:                 fmt.Sprintf(`{"emails":["homoglyph@ory.sh"],"username":"%cdmin"}`, 0x0430),
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"homoglyph@ory.sh", fmt.Sprintf("%cdmin", 0x0430)},
			ct:                  identity.CredentialsTypePassword,
		},
		// Two emails that differ only by a confusable character — Latin "a"
		// versus Cyrillic "а" (U+0430) — are not folded by NFKC, so both are
		// taken as distinct identifiers rather than collapsing into one.
		{
			doc:                 fmt.Sprintf(`{"emails":["admin@ory.sh","%cdmin@ory.sh"],"username":"confusable-emails"}`, 0x0430),
			schema:              "file://./stub/extension/credentials/multi.schema.json",
			expectedIdentifiers: []string{"admin@ory.sh", "аdmin@ory.sh", "confusable-emails"},
			ct:                  identity.CredentialsTypePassword,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			c := jsonschema.NewCompiler()
			runner, err := schema.NewExtensionRunner(t.Context())
			require.NoError(t, err)

			i := new(identity.Identity)
			e := identity.NewSchemaExtensionCredentials(i)
			if tc.existing != nil {
				i.SetCredentials(tc.ct, *tc.existing)
			}

			runner.AddRunner(e).Register(c)
			err = c.MustCompile(t.Context(), tc.schema).Validate(bytes.NewBufferString(tc.doc))
			if tc.expectErr != nil {
				require.EqualError(t, err, tc.expectErr.Error())
				return
			}
			require.NoError(t, err)
			require.NoError(t, e.Finish())

			credentials, ok := i.GetCredentials(tc.ct)
			require.True(t, ok)
			assert.ElementsMatch(t, tc.expectedIdentifiers, credentials.Identifiers)
			snapshotx.SnapshotT(t, credentials, snapshotx.ExceptPaths("identifiers"))
		})
	}
}
