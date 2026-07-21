// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/x/decoderx"
)

// decodeSubmission drives the production decode path used by the link
// strategy's decodeRecovery / decodeVerification.
func decodeSubmission(t *testing.T, schema []byte, email string) error {
	t.Helper()

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(schema)
	require.NoError(t, err)

	form := url.Values{"email": {email}, "method": {"link"}}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var body struct {
		Email string `json:"email"`
	}
	return decoderx.Decode(req, &body, compiler,
		decoderx.HTTPDecoderUseQueryAndBody(),
		decoderx.HTTPKeepRequestBody(true),
		decoderx.HTTPDecoderAllowedMethods("POST"),
		decoderx.HTTPDecoderSetValidatePayloads(true),
		decoderx.HTTPDecoderJSONFollowsFormFormat(),
	)
}

// TestLinkRecoverySubmissionAcceptsNonRFCEmail and
// TestLinkVerificationSubmissionAcceptsNonRFCEmail verify the `email_loose`
// submission schemas accept schema-approved non-RFC carrier addresses while
// still rejecting injection-unsafe input at decode.
func TestLinkRecoverySubmissionAcceptsNonRFCEmail(t *testing.T) {
	t.Parallel()

	require.NoError(t, decodeSubmission(t, recoveryMethodSchema, "user@example.org"))
	require.NoError(t, decodeSubmission(t, recoveryMethodSchema, "foo.@docomo.ne.jp"))
	for _, evil := range []string{"foo@ory.sh\r\nBcc: evil@x", "foo<bar>@ory.sh", "foo@ory.sh,evil@x"} {
		t.Run("rejected="+evil, func(t *testing.T) {
			require.Error(t, decodeSubmission(t, recoveryMethodSchema, evil))
		})
	}
}

func TestLinkVerificationSubmissionAcceptsNonRFCEmail(t *testing.T) {
	t.Parallel()

	require.NoError(t, decodeSubmission(t, verificationMethodSchema, "user@example.org"))
	require.NoError(t, decodeSubmission(t, verificationMethodSchema, "foo.@docomo.ne.jp"))
	for _, evil := range []string{"foo@ory.sh\r\nBcc: evil@x", "foo<bar>@ory.sh", "foo@ory.sh,evil@x"} {
		t.Run("rejected="+evil, func(t *testing.T) {
			require.Error(t, decodeSubmission(t, verificationMethodSchema, evil))
		})
	}
}
