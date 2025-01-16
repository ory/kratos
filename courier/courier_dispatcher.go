// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/x/otelx"
)

func (c *courier) channels(ctx context.Context, id string) (Channel, error) {
	cs, err := c.deps.CourierConfig().CourierChannels(ctx)
	if err != nil {
		return nil, err
	}

	for _, channel := range cs {
		if channel.ID != id {
			continue
		}
		switch channel.Type {
		case "smtp":
			courierChannel, err := NewSMTPChannelWithCustomTemplates(c.deps, channel.SMTPConfig, c.newEmailTemplateFromMessage)
			if err != nil {
				return nil, err
			}
			return courierChannel, nil
		case "http":
			return newHttpChannel(channel.ID, channel.RequestConfig, c.deps), nil
		default:
			return nil, errors.Errorf("unknown courier channel type: %s", channel.Type)
		}
	}

	return nil, errors.Errorf("no courier channels configured")
}

func (c *courier) DispatchMessage(ctx context.Context, msg Message) (err error) {
	ctx, span := c.deps.Tracer(ctx).Tracer().Start(ctx, "courier.DispatchMessage", trace.WithAttributes(
		attribute.Stringer("message.id", msg.ID),
		attribute.Stringer("message.nid", msg.NID),
		attribute.Stringer("message.type", msg.Type),
		attribute.String("message.template_type", string(msg.TemplateType)),
		attribute.Int("message.send_count", msg.SendCount),
	))
	defer otelx.End(span, &err)

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

	channel, err := c.channels(ctx, msg.Channel.String())
	if err != nil {
		return err
	}

	span.SetAttributes(attribute.String("channel.id", channel.ID()))
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

	//nolint:gosec // disable G115
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
			logger.
				WithError(err).
				Warn(`Unable to dispatch message.`)
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
