// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/courier/template"
	templates "github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
)

func queueNewMessage(t *testing.T, ctx context.Context, c courier.Courier, d template.Dependencies) uuid.UUID {
	t.Helper()
	id, err := c.QueueEmail(ctx, templates.NewTestStub(d, &templates.TestStubModel{
		To:      "test-recipient-1@example.org",
		Subject: "test-subject-1",
		Body:    "test-body-1",
	}))
	require.NoError(t, err)
	return id
}

func TestDispatchMessageWithInvalidSMTP(t *testing.T) {
	ctx := context.Background()

	conf, reg := internal.NewRegistryDefaultWithDSN(t, "")
	conf.MustSet(ctx, config.ViperKeyCourierMessageRetries, 5)
	conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "http://foo.url")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c, err := reg.Courier(ctx)
	require.NoError(t, err)

	t.Run("case=failed sending", func(t *testing.T) {
		id := queueNewMessage(t, ctx, c, reg)
		message, err := reg.CourierPersister().LatestQueuedMessage(ctx)
		require.NoError(t, err)
		require.Equal(t, id, message.ID)

		err = c.DispatchMessage(ctx, *message)
		// sending the email fails, because there is no SMTP server at foo.url
		require.Error(t, err)

		messages, err := reg.CourierPersister().NextMessages(ctx, 10)
		require.NoError(t, err)
		require.Len(t, messages, 1)
	})
}

func TestDispatchQueue(t *testing.T) {
	ctx := context.Background()

	conf, reg := internal.NewRegistryDefaultWithDSN(t, "")
	conf.MustSet(ctx, config.ViperKeyCourierMessageRetries, 1)

	c, err := reg.Courier(ctx)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	id := queueNewMessage(t, ctx, c, reg)
	require.NotEqual(t, uuid.Nil, id)

	// Fails to deliver the first time
	err = c.DispatchQueue(ctx)
	require.Error(t, err)

	// Retry once, as we set above - still fails
	err = c.DispatchQueue(ctx)
	require.Error(t, err)

	// Now it has been retried once, which means 2 > 1 is true and it is no longer tried
	err = c.DispatchQueue(ctx)
	require.NoError(t, err)

	var message courier.Message
	err = reg.Persister().GetConnection(ctx).
		Where("status = ?", courier.MessageStatusAbandoned).
		Eager("Dispatches").
		First(&message)

	require.NoError(t, err)
	require.Equal(t, id, message.ID)

	require.Len(t, message.Dispatches, 2)
	require.Contains(t, gjson.GetBytes(message.Dispatches[0].Error, "reason").String(), "failed to send email via smtp")
	require.Contains(t, gjson.GetBytes(message.Dispatches[1].Error, "reason").String(), "failed to send email via smtp")
}
