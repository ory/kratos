// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"

	"github.com/pkg/errors"
)

func (c *courier) DispatchMessage(ctx context.Context, msg Message) error {
	if err := c.deps.CourierPersister().IncrementMessageSendCount(ctx, msg.ID); err != nil {
		c.deps.Logger().
			WithError(err).
			WithField("message_id", msg.ID).
			WithField("message_nid", msg.NID).
			Error(`Unable to increment the message's "send_count" field`)
		return err
	}

	switch msg.Type {
	case MessageTypeEmail:
		if err := c.dispatchEmail(ctx, msg); err != nil {
			return err
		}
	case MessageTypePhone:
		if err := c.dispatchSMS(ctx, msg); err != nil {
			return err
		}
	default:
		return errors.Errorf("received unexpected message type: %d", msg.Type)
	}

	if err := c.deps.CourierPersister().SetMessageStatus(ctx, msg.ID, MessageStatusSent); err != nil {
		c.deps.Logger().
			WithError(err).
			WithField("message_id", msg.ID).
			WithField("message_nid", msg.NID).
			Error(`Unable to set the message status to "sent".`)
		return err
	}

	c.deps.Logger().
		WithField("message_id", msg.ID).
		WithField("message_nid", msg.NID).
		WithField("message_type", msg.Type).
		WithField("message_template_type", msg.TemplateType).
		WithField("message_subject", msg.Subject).
		Debug("Courier sent out message.")

	return nil
}

func (c *courier) DispatchQueue(ctx context.Context) error {
	maxRetries := c.deps.CourierConfig().CourierMessageRetries(ctx)
	pullCount := c.deps.CourierConfig().CourierWorkerPullCount(ctx)

	messages, err := c.deps.CourierPersister().NextMessages(ctx, uint8(pullCount))
	if err != nil {
		return err
	}

	for _, msg := range messages {
		if msg.SendCount > maxRetries {
			if err := c.deps.CourierPersister().SetMessageStatus(ctx, msg.ID, MessageStatusAbandoned); err != nil {
				c.deps.Logger().
					WithError(err).
					WithField("message_id", msg.ID).
					WithField("message_nid", msg.NID).
					Error(`Unable to set the retried message's status to "abandoned".`)
				return err
			}

			// Skip the message
			c.deps.Logger().
				WithField("message_id", msg.ID).
				WithField("message_nid", msg.NID).
				Warnf(`Message was abandoned because it did not deliver after %d attempts`, msg.SendCount)
			continue
		}

		if err := c.DispatchMessage(ctx, msg); err != nil {
			if err := c.deps.CourierPersister().RecordDispatch(ctx, msg.ID, CourierMessageDispatchStatusFailed, err); err != nil {
				c.deps.Logger().
					WithError(err).
					WithField("message_id", msg.ID).
					WithField("message_nid", msg.NID).
					Error(`Unable to record failure log entry.`)
				return err
			}

			if c.failOnDispatchError {
				return err
			}
			// an error happened, but we want to ignore it
			continue
		}

		if err := c.deps.CourierPersister().RecordDispatch(ctx, msg.ID, CourierMessageDispatchStatusSuccess, nil); err != nil {
			c.deps.Logger().
				WithError(err).
				WithField("message_id", msg.ID).
				WithField("message_nid", msg.NID).
				Error(`Unable to record success log entry.`)
			return err
		}
	}

	return nil
}
