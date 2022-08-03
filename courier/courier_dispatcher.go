package courier

import (
	"context"

	"github.com/pkg/errors"
)

func (c *courier) DispatchMessage(ctx context.Context, msg Message) error {
	maxRetries := c.deps.CourierConfig(ctx).CourierMessageRetries()

	if msg.SendCount > maxRetries {
		if err := c.deps.CourierPersister().SetMessageStatus(ctx, msg.ID, MessageStatusAbandoned); err != nil {
			c.deps.Logger().
				WithError(err).
				WithField("message_id", msg.ID).
				Error(`Unable to reset the retried message's status to "abandoned".`)
			return err
		}

		// Skip the message
		c.deps.Logger().
			WithField("message_id", msg.ID).
			Warnf(`Message was abandoned because it did not deliver after %d attempts`, msg.SendCount)
		return nil
	}

	if err := c.deps.CourierPersister().IncrementMessageSendCount(ctx, msg.ID); err != nil {
		c.deps.Logger().
			WithError(err).
			WithField("message_id", msg.ID).
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
			Error(`Unable to set the message status to "sent".`)
		return err
	}

	c.deps.Logger().
		WithField("message_id", msg.ID).
		WithField("message_type", msg.Type).
		WithField("message_template_type", msg.TemplateType).
		WithField("message_subject", msg.Subject).
		Debug("Courier sent out message.")

	return nil
}

func (c *courier) DispatchQueue(ctx context.Context) error {
	messages, err := c.deps.CourierPersister().NextMessages(ctx, 10)
	if err != nil {
		if errors.Is(err, ErrQueueEmpty) {
			return nil
		}
		return err
	}

	for k, msg := range messages {
		if err := c.DispatchMessage(ctx, msg); err != nil {

			for _, replace := range messages[k:] {
				if err := c.deps.CourierPersister().SetMessageStatus(ctx, replace.ID, MessageStatusQueued); err != nil {
					if c.failOnError {
						return err
					}
					c.deps.Logger().
						WithError(err).
						WithField("message_id", replace.ID).
						Error(`Unable to reset the failed message's status to "queued".`)
				}
			}

			return err
		}
	}

	return nil
}
