// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/flow/verification"
	"github.com/ory/kratos/selfservice/strategy/link"
	"github.com/ory/x/urlx"
)

func TestManager(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, conf)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/default.schema.json")
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh/")
	conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")
	conf.MustSet(ctx, config.ViperKeyLinkBaseURL, "https://link-url/")
	conf.MustSet(ctx, config.ViperKeySelfServiceRecoveryNotifyUnknownRecipients, true)
	conf.MustSet(ctx, config.ViperKeySelfServiceVerificationNotifyUnknownRecipients, true)

	u := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}

	i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
	i.Traits = identity.Traits(`{"email": "tracked@ory.sh"}`)
	require.NoError(t, reg.IdentityManager().Create(context.Background(), i))

	t.Run("method=SendRecoveryLink", func(t *testing.T) {
		s, err := reg.RecoveryStrategies(ctx).Strategy("link")
		require.NoError(t, err)
		f, err := recovery.NewFlow(conf, time.Hour, "", u, s, flow.TypeBrowser)
		require.NoError(t, err)

		require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(context.Background(), f))

		require.NoError(t, reg.LinkSender().SendRecoveryLink(context.Background(), f, "email", "tracked@ory.sh"))
		require.EqualError(t, reg.LinkSender().SendRecoveryLink(context.Background(), f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())

		messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
		require.NoError(t, err)
		require.Len(t, messages, 2)

		assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
		assert.Contains(t, messages[0].Subject, "Recover access to your account")
		assert.Contains(t, messages[0].Body, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(ctx), recovery.RouteSubmitFlow).String()+"?")
		assert.Contains(t, messages[0].Body, "token=")
		assert.Contains(t, messages[0].Body, "flow=")

		assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
		assert.Contains(t, messages[1].Subject, "Account access attempted")
		assert.NotContains(t, messages[1].Body, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(ctx), recovery.RouteSubmitFlow).String()+"?")
		assert.NotContains(t, messages[1].Body, "token=")
		assert.NotContains(t, messages[1].Body, "flow=")
	})

	t.Run("method=SendRecoveryLink via HTTP", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(2)
		type requestBody struct {
			Recipient    string
			RecoveryURL  string `json:"recovery_url"`
			To           string
			TemplateType string
			Subject      string
		}
		var messages []*requestBody
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var message requestBody
			require.NoError(t, json.Unmarshal(b, &message))
			messages = append(messages, &message)
			wg.Done()
		}))
		t.Cleanup(srv.Close)
		requestConfig := fmt.Sprintf(`{"url": "%s"}`, srv.URL)
		conf.MustSet(ctx, config.ViperKeyCourierDeliveryStrategy, "http")
		conf.MustSet(ctx, config.ViperKeyCourierHTTPRequestConfig, requestConfig)

		cour, err := reg.Courier(ctx)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(ctx)
		defer t.Cleanup(cancel)
		go func() {
			require.NoError(t, cour.Work(ctx))
		}()

		s, err := reg.RecoveryStrategies(ctx).Strategy("link")
		require.NoError(t, err)
		f, err := recovery.NewFlow(conf, time.Hour, "", u, s, flow.TypeBrowser)
		require.NoError(t, err)

		require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(context.Background(), f))

		require.NoError(t, reg.LinkSender().SendRecoveryLink(context.Background(), f, "email", "tracked@ory.sh"))
		require.EqualError(t, reg.LinkSender().SendRecoveryLink(context.Background(), f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())

		wg.Wait()

		assert.EqualValues(t, "tracked@ory.sh", messages[0].To)
		assert.Contains(t, messages[0].Subject, "Recover access to your account")
		assert.Contains(t, messages[0].RecoveryURL, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(ctx), recovery.RouteSubmitFlow).String()+"?")

		assert.EqualValues(t, "not-tracked@ory.sh", messages[1].To)
		assert.Contains(t, messages[1].Subject, "Account access attempted")
		assert.NotContains(t, messages[1].RecoveryURL, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(ctx), recovery.RouteSubmitFlow).String()+"?")
	})

	t.Run("method=SendVerificationLink", func(t *testing.T) {
		strategy, err := reg.GetActiveVerificationStrategy(ctx)
		require.NoError(t, err)

		f, err := verification.NewFlow(conf, time.Hour, "", u, strategy, flow.TypeBrowser)
		require.NoError(t, err)

		require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(context.Background(), f))

		require.NoError(t, reg.LinkSender().SendVerificationLink(context.Background(), f, "email", "tracked@ory.sh"))
		require.EqualError(t, reg.LinkSender().SendVerificationLink(context.Background(), f, "email", "not-tracked@ory.sh"), link.ErrUnknownAddress.Error())
		messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
		require.NoError(t, err)
		require.Len(t, messages, 2)

		assert.EqualValues(t, "tracked@ory.sh", messages[0].Recipient)
		assert.Contains(t, messages[0].Subject, "Please verify")
		assert.Contains(t, messages[0].Body, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(ctx), verification.RouteSubmitFlow).String()+"?")
		assert.Contains(t, messages[0].Body, "token=")
		assert.Contains(t, messages[0].Body, "flow=")

		assert.EqualValues(t, "not-tracked@ory.sh", messages[1].Recipient)
		assert.Contains(t, messages[1].Subject, "tried to verify")
		assert.NotContains(t, messages[1].Body, urlx.AppendPaths(conf.SelfServiceLinkMethodBaseURL(ctx), verification.RouteSubmitFlow).String()+"?")
		address, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, "tracked@ory.sh")
		require.NoError(t, err)
		assert.EqualValues(t, identity.VerifiableAddressStatusSent, address.Status)
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
					s, err := reg.RecoveryStrategies(ctx).Strategy("link")
					require.NoError(t, err)
					f, err := recovery.NewFlow(conf, time.Hour, "", u, s, flow.TypeBrowser)
					require.NoError(t, err)

					require.NoError(t, reg.RecoveryFlowPersister().CreateRecoveryFlow(context.Background(), f))

					err = reg.LinkSender().SendRecoveryLink(context.Background(), f, "email", "not-tracked@ory.sh")
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

					require.NoError(t, reg.VerificationFlowPersister().CreateVerificationFlow(context.Background(), f))

					err = reg.LinkSender().SendVerificationLink(context.Background(), f, "email", "not-tracked@ory.sh")
					require.ErrorIs(t, err, link.ErrUnknownAddress)
				},
			},
		} {
			t.Run("strategy="+tc.flow, func(t *testing.T) {

				conf.Set(ctx, tc.configKey, false)

				t.Cleanup(func() {
					conf.Set(ctx, tc.configKey, true)
				})

				tc.send(t)

				messages, err := reg.CourierPersister().NextMessages(context.Background(), 0)

				require.ErrorIs(t, err, courier.ErrQueueEmpty)
				require.Len(t, messages, 0)
			})
		}
	})
}
