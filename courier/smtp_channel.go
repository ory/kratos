// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"net"
	"net/textproto"
	"strconv"
	"time"

	"github.com/pkg/errors"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/herodot"
	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/mail/v3"
	"github.com/ory/x/otelx"
)

type (
	SMTPChannel struct {
		smtpClient *SMTPClient
		d          Dependencies

		newEmailTemplateFromMessage func(d template.Dependencies, msg Message) (EmailTemplate, error)
	}
)

var _ Channel = new(SMTPChannel)

func NewSMTPChannel(deps Dependencies, cfg *config.SMTPConfig) (*SMTPChannel, error) {
	return NewSMTPChannelWithCustomTemplates(deps, cfg, NewEmailTemplateFromMessage)
}

func NewSMTPChannelWithCustomTemplates(deps Dependencies, cfg *config.SMTPConfig, newEmailTemplateFromMessage func(d template.Dependencies, msg Message) (EmailTemplate, error)) (*SMTPChannel, error) {
	smtpClient, err := NewSMTPClient(deps, cfg)
	if err != nil {
		return nil, err
	}
	return &SMTPChannel{
		smtpClient:                  smtpClient,
		d:                           deps,
		newEmailTemplateFromMessage: newEmailTemplateFromMessage,
	}, nil
}

func (c *SMTPChannel) ID() string {
	return "email"
}

func (c *SMTPChannel) Dispatch(ctx context.Context, msg Message) (err error) {
	ctx, span := c.d.Tracer(ctx).Tracer().Start(ctx, "courier.SMTPChannel.Dispatch")
	defer otelx.End(span, &err)

	if c.smtpClient.Host == "" {
		return errors.WithStack(herodot.ErrInternalServerError.WithErrorf("Courier tried to deliver an email but %s is not set!", config.ViperKeyCourierSMTPURL))
	}

	channels, err := c.d.CourierConfig().CourierChannels(ctx)
	if err != nil {
		return err
	}

	var cfg *config.SMTPConfig
	for _, channel := range channels {
		if channel.ID == "email" && channel.SMTPConfig != nil {
			cfg = channel.SMTPConfig
			break
		}
	}

	if cfg == nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithErrorf("Courier tried to deliver an email but SMTP channel is misconfigured."))
	}

	gm := mail.NewMessage()
	if cfg.FromName == "" {
		gm.SetHeader("From", cfg.FromAddress)
	} else {
		gm.SetAddressHeader("From", cfg.FromAddress, cfg.FromName)
	}

	gm.SetHeader("To", msg.Recipient)
	gm.SetHeader("Subject", msg.Subject)

	headers := cfg.Headers
	for k, v := range headers {
		gm.SetHeader(k, v)
	}

	gm.SetBody("text/plain", msg.Body)

	logger := c.d.Logger().
		WithField("smtp_server", net.JoinHostPort(c.smtpClient.Host, strconv.Itoa(c.smtpClient.Port))).
		WithField("smtp_ssl_enabled", c.smtpClient.SSL).
		WithField("message_from", cfg.FromAddress).
		WithField("message_id", msg.ID).
		WithField("message_nid", msg.NID).
		WithField("message_type", msg.Type).
		WithField("message_template_type", msg.TemplateType).
		WithField("message_subject", msg.Subject).
		WithField("trace_id", span.SpanContext().TraceID())

	tmpl, err := c.newEmailTemplateFromMessage(c.d, msg)
	if err != nil {
		logger.
			WithError(err).Error(`Unable to get email template from message.`)
	} else if htmlBody, err := tmpl.EmailBody(ctx); err != nil {
		logger.
			WithError(err).Error(`Unable to get email body from template.`)
	} else {
		gm.AddAlternative("text/html", htmlBody)
	}

	dialCtx, dialSpan := c.d.Tracer(ctx).Tracer().Start(ctx, "courier.SMTPChannel.Dispatch.Dial", trace.WithAttributes(
		semconv.NetPeerName(c.smtpClient.Host),
		semconv.NetPeerPort(c.smtpClient.Port),
		semconv.NetProtocolName("smtp"),
	))
	snd, err := c.smtpClient.Dial(dialCtx)
	otelx.End(dialSpan, &err)

	if err != nil {
		logger.
			WithError(err).
			WithField("smtp_host", c.smtpClient.Host).
			WithField("smtp_port", c.smtpClient.Port).
			Error("Unable to dial SMTP connection.")
		return errors.WithStack(herodot.ErrInternalServerError.
			WithError(err.Error()).WithReason("failed to send email via smtp"))
	}
	defer func() { _ = snd.Close() }()

	sendCtx, sendSpan := c.d.Tracer(ctx).Tracer().Start(ctx, "courier.SMTPChannel.Dispatch.Send")
	err = mail.Send(sendCtx, snd, gm)
	otelx.End(sendSpan, &err)

	if err != nil {
		logger.
			WithError(err).
			Error("Unable to send email using SMTP connection.")

		var protoErr *textproto.Error
		var mailErr *mail.SendError

		switch {
		case errors.As(err, &mailErr) && errors.As(mailErr.Cause, &protoErr) && protoErr.Code >= 500:
			fallthrough
		case errors.As(err, &protoErr) && protoErr.Code >= 500:
			// See https://en.wikipedia.org/wiki/List_of_SMTP_server_return_codes
			// If the SMTP server responds with 5xx, sending the message should not be retried (without changing something about the request)
			if err := c.d.CourierPersister().SetMessageStatus(ctx, msg.ID, MessageStatusAbandoned); err != nil {
				logger.
					WithError(err).
					Error(`Unable to reset the retried message's status to "abandoned".`)
				return errors.WithStack(err)
			}
		}

		return errors.WithStack(herodot.ErrInternalServerError.
			WithError(err.Error()).WithReason("failed to send email via smtp"))
	}

	dispatchDuration := time.Since(msg.CreatedAt).Milliseconds()
	logger.WithField("dispatch_duration_ms", dispatchDuration).Debug("Courier sent out message.")

	return nil
}
