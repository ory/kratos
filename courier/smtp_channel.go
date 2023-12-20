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

func (c *SMTPChannel) Dispatch(ctx context.Context, msg Message) error {
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
			WithField("message_from", cfg.FromAddress).
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
