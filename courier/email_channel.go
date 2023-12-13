// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"fmt"
	"net/textproto"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/mail/v3"
)

type (
	EmailChannel struct {
		smtpClient  *SMTPClient
		httpChannel *httpChannel
		d           Dependencies

		newEmailTemplateFromMessage func(d template.Dependencies, msg Message) (EmailTemplate, error)
	}
)

var _ Channel = new(EmailChannel)

func NewEmailChannel(ctx context.Context, deps Dependencies) (*EmailChannel, error) {
	return NewEmailChannelWithCustomTemplates(ctx, deps, NewEmailTemplateFromMessage)
}

func NewEmailChannelWithCustomTemplates(ctx context.Context, deps Dependencies, newEmailTemplateFromMessage func(d template.Dependencies, msg Message) (EmailTemplate, error)) (*EmailChannel, error) {
	smtpClient, err := NewSMTP(ctx, deps)
	if err != nil {
		return nil, err
	}
	return &EmailChannel{
		smtpClient:                  smtpClient,
		httpChannel:                 newHttpChannel("emailViaHTTP", deps.CourierConfig().CourierEmailRequestConfig(ctx), deps),
		d:                           deps,
		newEmailTemplateFromMessage: newEmailTemplateFromMessage,
	}, nil
}

func (c *EmailChannel) ID() string {
	return "email"
}

func (c *EmailChannel) Dispatch(ctx context.Context, msg Message) error {
	if c.d.CourierConfig().CourierEmailStrategy(ctx) == "http" {
		return c.httpChannel.Dispatch(ctx, msg)
	}

	if c.smtpClient.Host == "" {
		return errors.WithStack(herodot.ErrInternalServerError.WithErrorf("Courier tried to deliver an email but %s is not set!", config.ViperKeyCourierSMTPURL))
	}

	from := c.d.CourierConfig().CourierSMTPFrom(ctx)
	fromName := c.d.CourierConfig().CourierSMTPFromName(ctx)

	gm := mail.NewMessage()
	if fromName == "" {
		gm.SetHeader("From", from)
	} else {
		gm.SetAddressHeader("From", from, fromName)
	}

	gm.SetHeader("To", msg.Recipient)
	gm.SetHeader("Subject", msg.Subject)

	headers := c.d.CourierConfig().CourierSMTPHeaders(ctx)
	for k, v := range headers {
		gm.SetHeader(k, v)
	}

	gm.SetBody("text/plain", msg.Body)

	tmpl, err := c.newEmailTemplateFromMessage(c.d, msg)
	if err != nil {
		c.d.Logger().
			WithError(err).
			WithField("message_id", msg.ID).
			WithField("message_nid", msg.NID).
			Error(`Unable to get email template from message.`)
	} else if htmlBody, err := tmpl.EmailBody(ctx); err != nil {
		c.d.Logger().
			WithError(err).
			WithField("message_id", msg.ID).
			WithField("message_nid", msg.NID).
			Error(`Unable to get email body from template.`)
	} else {
		gm.AddAlternative("text/html", htmlBody)
	}

	if err := c.smtpClient.DialAndSend(ctx, gm); err != nil {
		c.d.Logger().
			WithError(err).
			WithField("smtp_server", fmt.Sprintf("%s:%d", c.smtpClient.Host, c.smtpClient.Port)).
			WithField("smtp_ssl_enabled", c.smtpClient.SSL).
			WithField("message_from", from).
			WithField("message_id", msg.ID).
			WithField("message_nid", msg.NID).
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
				c.d.Logger().
					WithError(err).
					WithField("message_id", msg.ID).
					WithField("message_nid", msg.NID).
					Error(`Unable to reset the retried message's status to "abandoned".`)
				return err
			}
		}

		return errors.WithStack(herodot.ErrInternalServerError.
			WithError(err.Error()).WithReason("failed to send email via smtp"))
	}

	c.d.Logger().
		WithField("message_id", msg.ID).
		WithField("message_nid", msg.NID).
		WithField("message_type", msg.Type).
		WithField("message_template_type", msg.TemplateType).
		WithField("message_subject", msg.Subject).
		Debug("Courier sent out message.")

	return nil
}
