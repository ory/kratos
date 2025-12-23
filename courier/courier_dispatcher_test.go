// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier_test

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/courier"
	templates "github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/x/configx"
)

func queueNewMessage(t *testing.T, c courier.Courier) uuid.UUID {
	t.Helper()
	id, err := c.QueueEmail(t.Context(), templates.NewTestStub(&templates.TestStubModel{
		To:      "test-recipient-1@example.org",
		Subject: "test-subject-1",
		Body:    "test-body-1",
	}))
	require.NoError(t, err)
	return id
}

func TestDispatchMessageWithInvalidSMTP(t *testing.T) {
	_, reg := internal.NewRegistryDefaultWithDSN(t, "", configx.WithValues(map[string]any{
		config.ViperKeyCourierMessageRetries: 5,
		config.ViperKeyCourierSMTPURL:        "http://foo.url",
	}))

	c, err := reg.Courier(t.Context())
	require.NoError(t, err)

	t.Run("case=failed sending", func(t *testing.T) {
		id := queueNewMessage(t, c)
		message, err := reg.CourierPersister().LatestQueuedMessage(t.Context())
		require.NoError(t, err)
		require.Equal(t, id, message.ID)

		err = c.DispatchMessage(t.Context(), *message)
		// sending the email fails, because there is no SMTP server at foo.url
		require.Error(t, err)

		messages, err := reg.CourierPersister().NextMessages(t.Context(), 10)
		require.NoError(t, err)
		require.Len(t, messages, 1)
	})
}

func TestDispatchMessage(t *testing.T) {
	_, reg := internal.NewRegistryDefaultWithDSN(t, "", configx.WithValues(map[string]any{
		config.ViperKeyCourierMessageRetries: 5,
		config.ViperKeyCourierSMTPURL:        "http://foo.url",
	}))

	c, err := reg.Courier(t.Context())
	require.NoError(t, err)
	t.Run("case=invalid channel", func(t *testing.T) {
		message := courier.Message{
			Channel:      "invalid-channel",
			Status:       courier.MessageStatusQueued,
			Type:         courier.MessageTypeEmail,
			Recipient:    testhelpers.RandomEmail(),
			Subject:      "test-subject-1",
			Body:         "test-body-1",
			TemplateType: "stub",
		}
		require.NoError(t, reg.CourierPersister().AddMessage(t.Context(), &message))
		assert.ErrorContains(t, c.DispatchMessage(t.Context(), message), "no courier channels configured for: invalid-channel")
	})
}

func TestDispatchQueue(t *testing.T) {
	_, reg := internal.NewRegistryDefaultWithDSN(t, "", configx.WithValue(config.ViperKeyCourierMessageRetries, 1))

	c, err := reg.Courier(t.Context())
	require.NoError(t, err)
	c.FailOnDispatchError()

	id := queueNewMessage(t, c)
	require.NotEqual(t, uuid.Nil, id)

	// Fails to deliver the first time
	err = c.DispatchQueue(t.Context())
	require.Error(t, err)

	// Retry once, as we set above - still fails
	err = c.DispatchQueue(t.Context())
	require.Error(t, err)

	// Now it has been retried once, which means 2 > 1 is true and it is no longer tried
	err = c.DispatchQueue(t.Context())
	require.NoError(t, err)

	var message courier.Message
	err = reg.Persister().GetConnection(t.Context()).
		Where("status = ?", courier.MessageStatusAbandoned).
		Eager("Dispatches").
		First(&message)

	require.NoError(t, err)
	require.Equal(t, id, message.ID)

	require.Len(t, message.Dispatches, 2)
	require.Contains(t, gjson.GetBytes(message.Dispatches[0].Error, "reason").String(), "failed to send email via smtp")
	require.Contains(t, gjson.GetBytes(message.Dispatches[1].Error, "reason").String(), "failed to send email via smtp")
}
