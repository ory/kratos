// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"

	"github.com/pkg/errors"
)

func (c *courier) DispatchMessage(ctx context.Context, msg Message) error {
	logger := c.deps.Logger().
		WithField("message_id", msg.ID).
		WithField("message_nid", msg.NID).
		WithField("message_type", msg.Type).
		WithField("message_template_type", msg.TemplateType).
		WithField("message_subject", msg.Subject)

	if err := c.deps.CourierPersister().IncrementMessageSendCount(ctx, msg.ID); err != nil {
		logger.
			WithError(err).
			Error(`Unable to increment the message's "send_count" field`)
		return err
	}

	channel, ok := c.courierChannels[msg.Channel.String()]
	if !ok {
		return errors.Errorf("message %s has unknown channel %q", msg.ID.String(), msg.Channel)
	}

	logger = logger.
		WithField("channel", channel.ID())

	if err := channel.Dispatch(ctx, msg); err != nil {
		return err
	}

	if err := c.deps.CourierPersister().SetMessageStatus(ctx, msg.ID, MessageStatusSent); err != nil {
		logger.
			WithError(err).
			Error(`Unable to set the message status to "sent".`)
		return err
	}

	logger.Debug("Courier sent out message.")

	return nil
}

func (c *courier) DispatchQueue(ctx context.Context) error {
	maxRetries := c.deps.CourierConfig().CourierMessageRetries(ctx)
	pullCount := c.deps.CourierConfig().CourierWorkerPullCount(ctx)

	messages, err := c.deps.CourierPersister().NextMessages(ctx, uint8(pullCount))
	if err != nil {
		if errors.Is(err, ErrQueueEmpty) {
			return nil
		}
		return err
	}

	for k, msg := range messages {
		logger := c.deps.Logger().
			WithField("message_id", msg.ID).
			WithField("message_nid", msg.NID).
			WithField("message_type", msg.Type).
			WithField("message_template_type", msg.TemplateType).
			WithField("message_subject", msg.Subject)

		if msg.SendCount > maxRetries {
			if err := c.deps.CourierPersister().SetMessageStatus(ctx, msg.ID, MessageStatusAbandoned); err != nil {
				logger.
					WithError(err).
					Error(`Unable to set the retried message's status to "abandoned".`)
				return err
			}

			// Skip the message
			logger.
				Warnf(`Message was abandoned because it did not deliver after %d attempts`, msg.SendCount)
		} else if err := c.DispatchMessage(ctx, msg); err != nil {
			if err := c.deps.CourierPersister().RecordDispatch(ctx, msg.ID, CourierMessageDispatchStatusFailed, err); err != nil {
				logger.
					WithError(err).
					Error(`Unable to record failure log entry.`)
				if c.failOnDispatchError {
					return err
				}
			}

			for _, replace := range messages[k:] {
				if err := c.deps.CourierPersister().SetMessageStatus(ctx, replace.ID, MessageStatusQueued); err != nil {
					logger.
						WithError(err).
						Error(`Unable to reset the failed message's status to "queued".`)
					if c.failOnDispatchError {
						return err
					}
				}
			}

			if c.failOnDispatchError {
				return err
			}
		} else if err := c.deps.CourierPersister().RecordDispatch(ctx, msg.ID, CourierMessageDispatchStatusSuccess, nil); err != nil {
			logger.
				WithError(err).
				Error(`Unable to record success log entry.`)
			// continue with execution, as the message was successfully dispatched
		}
	}

	return nil
}
