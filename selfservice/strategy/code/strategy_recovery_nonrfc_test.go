// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/x/decoderx"
)

// decodeRecoverySubmission drives the production decode path used by
// (*Strategy).decodeRecovery.
func decodeRecoverySubmission(t *testing.T, email string) error {
	t.Helper()

	compiler, err := decoderx.HTTPRawJSONSchemaCompiler(recoveryMethodSchema)
	require.NoError(t, err)

	form := url.Values{"email": {email}, "method": {"code"}}
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

// TestCodeRecoverySubmissionAcceptsNonRFCEmail verifies the `email_loose`
// recovery submission schema accepts schema-approved non-RFC carrier addresses
// while still rejecting injection-unsafe input at decode.
func TestCodeRecoverySubmissionAcceptsNonRFCEmail(t *testing.T) {
	t.Parallel()

	require.NoError(t, decodeRecoverySubmission(t, "user@example.org"))
	require.NoError(t, decodeRecoverySubmission(t, "foo.@docomo.ne.jp"))

	for _, evil := range []string{"foo@ory.sh\r\nBcc: evil@x", "foo<bar>@ory.sh", "foo@ory.sh,evil@x"} {
		t.Run("rejected="+evil, func(t *testing.T) {
			require.Error(t, decodeRecoverySubmission(t, evil))
		})
	}
}
