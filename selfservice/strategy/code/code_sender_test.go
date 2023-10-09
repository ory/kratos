// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/x/urlx"
)

var b64 = func(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func TestSender(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/default.schema.json")
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh/")
	conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")
	conf.MustSet(ctx, config.ViperKeyLinkBaseURL, "https://link-url/")
	conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, true)
	conf.MustSet(ctx, config.ViperKeySelfServiceVerificationNotifyUnknownRecipients, true)

	u := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}

	i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Traits = identity.Traits(`{"email": "tracked@ory.sh"}`)
	require.NoError(t, reg.IdentityManager().Create(ctx, i))

	t.Run("method=SendRecoveryCode", func(t *testing.T) {
		recoveryCode := func(t *testing.T) {
			t.Helper()
			f, err := recovery.NewFlow(conf, time.Hour, "", u, code.NewStrategy(reg), flow.TypeBrowser)
			require.NoError(t, err)

			require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(ctx, f))

			require.NoError(t, reg.CodeSender().SendRecoveryCode(ctx, f, "email", "tracked@ory.sh"))
			require.ErrorIs(t, reg.CodeSender().SendRecoveryCode(ctx, f, "email", "not-tracked@ory.sh"), code.ErrUnknownAddress)
		}

		t.Run("case=with default templates", func(t *testing.T) {
			recoveryCode(t)
			messages, err := reg.CourierPersister().NextMessages(ctx, 12)
			require.NoError(t, err)
			require.Len(t, messages, 2)

			assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
			assert.Contains(t, messages[0].Subject, "Recover access to your account")

			assert.Regexp(t, testhelpers.CodeRegex, messages[0].Body)

			assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
			assert.Contains(t, messages[1].Subject, "Account access attempted")

			assert.NotRegexp(t, testhelpers.CodeRegex, messages[1].Body, "Expected message to not contain an 6 digit recovery code, but it did: ", messages[1].Body)
		})

		t.Run("case=with custom templates", func(t *testing.T) {
			subject := "custom template recovery code"
			body := "custom template recovery code body"
			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeyCourierTemplatesRecoveryCodeInvalidEmail, nil)
				conf.MustSet(ctx, config.ViperKeyCourierTemplatesRecoveryCodeValidEmail, nil)
			})
			conf.MustSet(ctx, config.ViperKeyCourierTemplatesRecoveryCodeInvalidEmail, fmt.Sprintf(`{ "subject": "base64://%s", "body": { "plaintext": "base64://%s", "html": "base64://%s" }}`, b64(subject+" invalid"), b64(body), b64(body)))
			conf.MustSet(ctx, config.ViperKeyCourierTemplatesRecoveryCodeValidEmail, fmt.Sprintf(`{ "subject": "base64://%s", "body": { "plaintext": "base64://%s", "html": "base64://%s" }}`, b64(subject+" valid"), b64(body+" {{ .RecoveryCode }}"), b64(body+" {{ .RecoveryCode }}")))
			recoveryCode(t)
			messages, err := reg.CourierPersister().NextMessages(ctx, 12)
			require.NoError(t, err)
			require.Len(t, messages, 2)

			assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
			assert.Equal(t, messages[0].Subject, subject+" valid")
			assert.Contains(t, messages[0].Body, body)

			assert.Regexp(t, testhelpers.CodeRegex, messages[0].Body)

			assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
			assert.Equal(t, messages[1].Subject, subject+" invalid")
			assert.Equal(t, messages[1].Body, body)
		})
	})

	t.Run("method=SendVerificationCode", func(t *testing.T) {
		verificationFlow := func(t *testing.T) {
			t.Helper()

			f, err := verification.NewFlow(conf, time.Hour, "", u, code.NewStrategy(reg), flow.TypeBrowser)
			require.NoError(t, err)

			require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

			require.NoError(t, reg.CodeSender().SendVerificationCode(ctx, f, "email", "tracked@ory.sh"))
			require.ErrorIs(t, reg.CodeSender().SendVerificationCode(ctx, f, "email", "not-tracked@ory.sh"), code.ErrUnknownAddress)
		}

		t.Run("case=with default templates", func(t *testing.T) {
			verificationFlow(t)
			messages, err := reg.CourierPersister().NextMessages(ctx, 12)
			require.NoError(t, err)
			require.Len(t, messages, 2)

			assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
			assert.Contains(t, messages[0].Subject, "Please verify your email address")

			assert.Regexp(t, testhelpers.CodeRegex, messages[0].Body)

			assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
			assert.Contains(t, messages[1].Subject, "Someone tried to verify this email address")

			assert.NotRegexp(t, testhelpers.CodeRegex, messages[1].Body, "Expected message to not contain an 6 digit recovery code, but it did: ", messages[1].Body)
		})

		t.Run("case=with custom templates", func(t *testing.T) {
			subject := "custom template verification code"
			body := "custom template verification code body"
			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeyCourierTemplatesVerificationCodeInvalidEmail, nil)
				conf.MustSet(ctx, config.ViperKeyCourierTemplatesVerificationCodeValidEmail, nil)
			})
			conf.MustSet(ctx, config.ViperKeyCourierTemplatesVerificationCodeInvalidEmail, fmt.Sprintf(`{ "subject": "base64://%s", "body": { "plaintext": "base64://%s", "html": "base64://%s" }}`, b64(subject+" invalid"), b64(body), b64(body)))
			conf.MustSet(ctx, config.ViperKeyCourierTemplatesVerificationCodeValidEmail, fmt.Sprintf(`{ "subject": "base64://%s", "body": { "plaintext": "base64://%s", "html": "base64://%s" }}`, b64(subject+" valid"), b64(body+" {{ .VerificationCode }}"), b64(body+" {{ .VerificationCode }}")))
			verificationFlow(t)
			messages, err := reg.CourierPersister().NextMessages(ctx, 12)
			require.NoError(t, err)
			require.Len(t, messages, 2)

			assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
			assert.Equal(t, messages[0].Subject, subject+" valid")
			assert.Contains(t, messages[0].Body, body)

			assert.Regexp(t, testhelpers.CodeRegex, messages[0].Body)

			assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
			assert.Equal(t, messages[1].Subject, subject+" invalid")
			assert.Equal(t, messages[1].Body, body)
		})
	})

	t.Run("case=should be able to disable invalid email dispatch", func(t *testing.T) {
		for _, tc := range []struct {
			flow      string
			send      func(t *testing.T)
			configKey string
		}{
			{
				flow:      "recovery",
				configKey: config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients,
				send: func(t *testing.T) {
					s, err := reg.RecoveryStrategies(ctx).Strategy("code")
					require.NoError(t, err)
					f, err := recovery.NewFlow(conf, time.Hour, "", u, s, flow.TypeBrowser)
					require.NoError(t, err)

					require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(ctx, f))

					err = reg.CodeSender().SendRecoveryCode(ctx, f, "email", "not-tracked@ory.sh")
					require.ErrorIs(t, err, code.ErrUnknownAddress)
				},
			},
			{
				flow:      "verification",
				configKey: config.ViperKeySelfServiceVerificationNotifyUnknownRecipients,
				send: func(t *testing.T) {
					s, err := reg.VerificationStrategies(ctx).Strategy("code")
					require.NoError(t, err)
					f, err := verification.NewFlow(conf, time.Hour, "", u, s, flow.TypeBrowser)
					require.NoError(t, err)

					require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

					err = reg.CodeSender().SendVerificationCode(ctx, f, "email", "not-tracked@ory.sh")
					require.ErrorIs(t, err, code.ErrUnknownAddress)
				},
			},
		} {
			t.Run("strategy="+tc.flow, func(t *testing.T) {
				conf.Set(ctx, tc.configKey, false)

				t.Cleanup(func() {
					conf.Set(ctx, tc.configKey, true)
				})

				tc.send(t)

				messages, err := reg.CourierPersister().NextMessages(ctx, 0)

				require.ErrorIs(t, err, courier.ErrQueueEmpty)
				require.Len(t, messages, 0)
			})
		}
	})
}
