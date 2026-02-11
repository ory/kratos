// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/pkg/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/kratos/x"
	"github.com/ory/x/contextx"
	"github.com/ory/x/urlx"
)

func TestManager(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	conf, reg := pkg.NewFastRegistryWithMocks(t)
	initViper(t, conf)
	ctx = testhelpers.WithDefaultIdentitySchema(ctx, "file://./stub/default.schema.json")
	ctx = contextx.WithConfigValues(ctx, map[string]any{
		config.ViperKeyPublicBaseURL:                                  "https://www.ory.sh/",
		config.ViperKeyCourierSMTPURL:                                 "smtp://foo@bar@dev.null/",
		config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients:     true,
		config.ViperKeySelfServiceVerificationNotifyUnknownRecipients: true,
	})

	u := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}

	i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Traits = identity.Traits(`{"email": "tracked@ory.sh"}`)
	require.NoError(t, reg.IdentityManager().Create(ctx, i))

	for _, tc := range []struct {
		d               string
		setupContext    func(ctx context.Context) context.Context
		recoveryURL     string
		verificationURL string
	}{{
		d:               "without BaseURL",
		setupContext:    func(ctx context.Context) context.Context { return ctx },
		recoveryURL:     "https://www.ory.sh/self-service/recovery?flow=",
		verificationURL: "https://www.ory.sh/self-service/verification?flow=",
	}, {
		d: "with BaseURL",
		setupContext: func(ctx context.Context) context.Context {
			return x.WithBaseURL(ctx, urlx.ParseOrPanic("https://proxy.example.com/some/subpath/"))
		},
		recoveryURL:     "https://proxy.example.com/some/subpath/self-service/recovery?flow=",
		verificationURL: "https://proxy.example.com/some/subpath/self-service/verification?flow=",
	}} {
		ctx := tc.setupContext(ctx)

		t.Run("case="+tc.d, func(t *testing.T) {
			t.Run("method=SendRecoveryLink", func(t *testing.T) {
				s, err := reg.RecoveryStrategies(ctx).Strategy("link")
				require.NoError(t, err)
				f, err := recovery.NewFlow(conf, time.Hour, "", u, s, flow.TypeBrowser)
				require.NoError(t, err)

				require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(ctx, f))

				require.NoError(t, reg.LinkSender().SendRecoveryLink(ctx, f, "email", "tracked@ory.sh"))
				require.EqualError(t, reg.LinkSender().SendRecoveryLink(ctx, f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())

				messages, err := reg.CourierPersister().NextMessages(ctx, 12)
				require.NoError(t, err)
				require.Len(t, messages, 2)

				assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
				assert.Contains(t, messages[0].Subject, "Recover access to your account")
				assert.Contains(t, messages[0].Body, tc.recoveryURL)
				assert.Contains(t, messages[0].Body, "token=")
				assert.Contains(t, messages[0].Body, "flow=")

				assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
				assert.Contains(t, messages[1].Subject, "Account access attempted")
				assert.NotContains(t, messages[1].Body, f.RequestURL+"self-service/recovery?flow=")
				assert.NotContains(t, messages[1].Body, "token=")
				assert.NotContains(t, messages[1].Body, "flow=")
			})

			t.Run("method=SendRecoveryLink via HTTP", func(t *testing.T) {
				type requestBody struct {
					Recipient    string
					RecoveryURL  string `json:"recovery_url"`
					To           string
					TemplateType string
					Subject      string
				}
				messages := make(chan *requestBody, 2)
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var message requestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&message))
					messages <- &message
				}))
				t.Cleanup(srv.Close)

				ctx = contextx.WithConfigValues(ctx, map[string]any{
					config.ViperKeyCourierDeliveryStrategy:           "http",
					config.ViperKeyCourierHTTPRequestConfig + ".url": srv.URL,
				})

				s, err := reg.RecoveryStrategies(ctx).Strategy("link")
				require.NoError(t, err)
				f, err := recovery.NewFlow(conf, time.Hour, "", u, s, flow.TypeBrowser)
				require.NoError(t, err)

				require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(ctx, f))

				require.NoError(t, reg.LinkSender().SendRecoveryLink(ctx, f, "email", "tracked@ory.sh"))
				require.EqualError(t, reg.LinkSender().SendRecoveryLink(ctx, f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())

				cour, err := reg.Courier(ctx)
				require.NoError(t, err)

				require.NoError(t, cour.DispatchQueue(ctx))
				close(messages)

				require.Len(t, messages, 2)

				msg := <-messages
				assert.EqualValues(t, "tracked@ory.sh", msg.To)
				assert.Contains(t, msg.Subject, "Recover access to your account")
				assert.Contains(t, msg.RecoveryURL, tc.recoveryURL)

				msg = <-messages
				assert.EqualValues(t, "not-tracked@ory.sh", msg.To)
				assert.Contains(t, msg.Subject, "Account access attempted")
				assert.NotContains(t, msg.RecoveryURL, tc.recoveryURL)
			})

			t.Run("method=SendVerificationLink", func(t *testing.T) {
				strategy, err := reg.GetActiveVerificationStrategy(ctx)
				require.NoError(t, err)

				f, err := verification.NewFlow(conf, time.Hour, "", u, strategy, flow.TypeBrowser)
				require.NoError(t, err)

				require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

				require.NoError(t, reg.LinkSender().SendVerificationLink(ctx, f, "email", "tracked@ory.sh"))
				require.EqualError(t, reg.LinkSender().SendVerificationLink(ctx, f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())
				messages, err := reg.CourierPersister().NextMessages(ctx, 12)
				require.NoError(t, err)
				require.Len(t, messages, 2)

				assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
				assert.Contains(t, messages[0].Subject, "Please verify")
				assert.Contains(t, messages[0].Body, tc.verificationURL)
				assert.Contains(t, messages[0].Body, "token=")
				assert.Contains(t, messages[0].Body, "flow=")

				assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
				assert.Contains(t, messages[1].Subject, "tried to verify")
				assert.NotContains(t, messages[1].Body, tc.verificationURL)
				address, err := reg.IdentityPool().FindVerifiableAddressByValue(ctx, identity.AddressTypeEmail, "tracked@ory.sh")
				require.NoError(t, err)
				assert.EqualValues(t, identity.VerifiableAddressStatusSent, address.Status)
			})
		})
	}

	t.Run("case=should be able to disable invalid email dispatch", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			flow      string
			send      func(t *testing.T)
			configKey string
		}{
			{
				flow:      "recovery",
				configKey: config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients,
				send: func(t *testing.T) {
					s, err := reg.RecoveryStrategies(ctx).Strategy("link")
					require.NoError(t, err)
					f, err := recovery.NewFlow(conf, time.Hour, "", u, s, flow.TypeBrowser)
					require.NoError(t, err)

					require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(ctx, f))

					err = reg.LinkSender().SendRecoveryLink(ctx, f, "email", "not-tracked@ory.sh")
					require.ErrorIs(t, err, link.ErrUnknownAddress)
				},
			},
			{
				flow:      "verification",
				configKey: config.ViperKeySelfServiceVerificationNotifyUnknownRecipients,
				send: func(t *testing.T) {
					s, err := reg.VerificationStrategies(ctx).Strategy("link")
					require.NoError(t, err)
					f, err := verification.NewFlow(conf, time.Hour, "", u, s, flow.TypeBrowser)
					require.NoError(t, err)

					require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(ctx, f))

					err = reg.LinkSender().SendVerificationLink(ctx, f, "email", "not-tracked@ory.sh")
					require.ErrorIs(t, err, link.ErrUnknownAddress)
				},
			},
		} {
			t.Run("strategy="+tc.flow, func(t *testing.T) {
				t.Parallel()

				ctx = contextx.WithConfigValue(ctx, tc.configKey, false)

				tc.send(t)

				messages, err := reg.CourierPersister().NextMessages(ctx, 0)

				require.ErrorIs(t, err, courier.ErrQueueEmpty)
				require.Len(t, messages, 0)
			})
		}
	})
}
