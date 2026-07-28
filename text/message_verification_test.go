// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/clock"
)

func TestNewErrorValidationVerificationFlowExpired(t *testing.T) {
	t.Parallel()

	expiredAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	// "Now" is exactly two minutes after the flow expired.
	c := clock.NewMock(expiredAt.Add(2 * time.Minute))

	msg := NewErrorValidationVerificationFlowExpired(c, expiredAt)

	assert.Equal(t, ErrorValidationVerificationFlowExpired, msg.ID)
	assert.Equal(t, "The verification flow expired 2.00 minutes ago, please try again.", msg.Text)
}

func TestVerificationCodeSentMessage(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		via    string
		wantID ID
	}{
		{via: "sms", wantID: InfoSelfServiceVerificationPhoneWithCodeSent},
		{via: "email", wantID: InfoSelfServiceVerificationEmailWithCodeSent},
		// Any unknown channel falls back to the email message.
		{via: "", wantID: InfoSelfServiceVerificationEmailWithCodeSent},
	} {
		t.Run("via="+tc.via, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.wantID, VerificationCodeSentMessage(tc.via).ID)
		})
	}
}
