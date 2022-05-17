package courier

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ory/kratos/courier/template"

	"github.com/ory/kratos/driver/config"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	gomail "github.com/ory/mail/v3"
)

type smtpClient struct {
	*gomail.Dialer

	GetTemplateType        func(t EmailTemplate) (TemplateType, error)
	NewTemplateFromMessage func(d template.Dependencies, msg Message) (EmailTemplate, error)
}

func newSMTP(ctx context.Context, deps Dependencies) *smtpClient {
	uri := deps.CourierConfig(ctx).CourierSMTPURL()
	var tlsCertificates []tls.Certificate
	clientCertPath := deps.CourierConfig(ctx).CourierSMTPClientCertPath()
	clientKeyPath := deps.CourierConfig(ctx).CourierSMTPClientKeyPath()

	if clientCertPath != "" && clientKeyPath != "" {
		clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err == nil {
			tlsCertificates = append(tlsCertificates, clientCert)
		} else {
			deps.Logger().
				WithError(err).
				Error("Unable to load tls certificate and private key for smtp client.")
		}
	}

	localName := deps.CourierConfig(ctx).CourierSMTPLocalName()
	password, _ := uri.User.Password()
	port, _ := strconv.ParseInt(uri.Port(), 10, 0)

	dialer := &gomail.Dialer{
		Host:      uri.Hostname(),
		Port:      int(port),
		Username:  uri.User.Username(),
		Password:  password,
		LocalName: localName,

		Timeout:      time.Second * 10,
		RetryFailure: true,
	}

	sslSkipVerify, _ := strconv.ParseBool(uri.Query().Get("skip_ssl_verify"))

	serverName := uri.Query().Get("server_name")
	if serverName == "" {
		serverName = uri.Hostname()
	}

	// SMTP schemes
	// smtp: smtp clear text (with uri parameter) or with StartTLS (enforced by default)
	// smtps: smtp with implicit TLS (recommended way in 2021 to avoid StartTLS downgrade attacks
	//    and defaulting to fully-encrypted protocols https://datatracker.ietf.org/doc/html/rfc8314)
	switch uri.Scheme {
	case "smtp":
		// Enforcing StartTLS by default for security best practices (config review, etc.)
		skipStartTLS, _ := strconv.ParseBool(uri.Query().Get("disable_starttls"))
		if !skipStartTLS {
			// #nosec G402 This is ok (and required!) because it is configurable and disabled by default.
			dialer.TLSConfig = &tls.Config{InsecureSkipVerify: sslSkipVerify, Certificates: tlsCertificates, ServerName: serverName}
			// Enforcing StartTLS
			dialer.StartTLSPolicy = gomail.MandatoryStartTLS
		}
	case "smtps":
		// #nosec G402 This is ok (and required!) because it is configurable and disabled by default.
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: sslSkipVerify, Certificates: tlsCertificates, ServerName: serverName}
		dialer.SSL = true
	}

	return &smtpClient{
		Dialer: dialer,

		GetTemplateType:        GetEmailTemplateType,
		NewTemplateFromMessage: NewEmailTemplateFromMessage,
	}
}

func (c *courier) SetGetEmailTemplateType(f func(t EmailTemplate) (TemplateType, error)) {
	c.smtpClient.GetTemplateType = f
}

func (c *courier) SetNewEmailTemplateFromMessage(f func(d template.Dependencies, msg Message) (EmailTemplate, error)) {
	c.smtpClient.NewTemplateFromMessage = f
}

func (c *courier) SmtpDialer() *gomail.Dialer {
	return c.smtpClient.Dialer
}

func (c *courier) QueueEmail(ctx context.Context, t EmailTemplate) (uuid.UUID, error) {
	recipient, err := t.EmailRecipient()
	if err != nil {
		return uuid.Nil, err
	}

	subject, err := t.EmailSubject(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	bodyPlaintext, err := t.EmailBodyPlaintext(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	templateType, err := c.smtpClient.GetTemplateType(t)
	if err != nil {
		return uuid.Nil, err
	}

	templateData, err := json.Marshal(t)
	if err != nil {
		return uuid.Nil, err
	}

	message := &Message{
		Status:       MessageStatusQueued,
		Type:         MessageTypeEmail,
		Recipient:    recipient,
		Body:         bodyPlaintext,
		Subject:      subject,
		TemplateType: templateType,
		TemplateData: templateData,
	}

	if err := c.deps.CourierPersister().AddMessage(ctx, message); err != nil {
		return uuid.Nil, err
	}

	return message.ID, nil
}

func (c *courier) dispatchEmail(ctx context.Context, msg Message) error {
	if c.smtpClient.Host == "" {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Courier tried to deliver an email but %s is not set!", config.ViperKeyCourierSMTPURL))
	}

	from := c.deps.CourierConfig(ctx).CourierSMTPFrom()
	fromName := c.deps.CourierConfig(ctx).CourierSMTPFromName()

	gm := gomail.NewMessage()
	if fromName == "" {
		gm.SetHeader("From", from)
	} else {
		gm.SetAddressHeader("From", from, fromName)
	}

	gm.SetHeader("To", msg.Recipient)
	gm.SetHeader("Subject", msg.Subject)

	headers := c.deps.CourierConfig(ctx).CourierSMTPHeaders()
	for k, v := range headers {
		gm.SetHeader(k, v)
	}

	gm.SetBody("text/plain", msg.Body)

	tmpl, err := c.smtpClient.NewTemplateFromMessage(c.deps, msg)
	if err != nil {
		c.deps.Logger().
			WithError(err).
			WithField("message_id", msg.ID).
			Error(`Unable to get email template from message.`)
	} else {
		htmlBody, err := tmpl.EmailBody(ctx)
		if err != nil {
			c.deps.Logger().
				WithError(err).
				WithField("message_id", msg.ID).
				Error(`Unable to get email body from template.`)
		} else {
			gm.AddAlternative("text/html", htmlBody)
		}
	}

	if err := c.smtpClient.DialAndSend(ctx, gm); err != nil {
		c.deps.Logger().
			WithError(err).
			WithField("smtp_server", fmt.Sprintf("%s:%d", c.smtpClient.Host, c.smtpClient.Port)).
			WithField("smtp_ssl_enabled", c.smtpClient.SSL).
			// WithField("email_to", msg.Recipient).
			WithField("message_from", from).
			Error("Unable to send email using SMTP connection.")
		return errors.WithStack(err)
	}

	c.deps.Logger().
		WithField("message_id", msg.ID).
		WithField("message_type", msg.Type).
		WithField("message_template_type", msg.TemplateType).
		WithField("message_subject", msg.Subject).
		Debug("Courier sent out message.")

	return nil
}
