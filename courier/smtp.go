// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/mail"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/kratos/driver/config"

	"github.com/gofrs/uuid"

	gomail "github.com/ory/mail/v3"
)

type SMTPClient struct {
	*gomail.Dialer
}

func NewSMTPClient(deps Dependencies, cfg *config.SMTPConfig) (*SMTPClient, error) {
	uri, err := url.Parse(cfg.ConnectionURI)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrMisconfiguration.WithReasonf("The SMTP connection URI is malformed. Please contact a system administrator."))
	}

	var tlsCertificates []tls.Certificate

	if cfg.ClientCertPath != "" && cfg.ClientKeyPath != "" {
		clientCert, err := tls.LoadX509KeyPair(cfg.ClientCertPath, cfg.ClientKeyPath)
		if err == nil {
			tlsCertificates = append(tlsCertificates, clientCert)
		} else {
			deps.Logger().
				WithError(err).
				Error("Unable to load tls certificate and private key for smtp client.")
		}
	}

	password, _ := uri.User.Password()
	port, _ := strconv.ParseInt(uri.Port(), 10, 0)

	dialer := &gomail.Dialer{
		Host:      uri.Hostname(),
		Port:      int(port),
		Username:  uri.User.Username(),
		Password:  password,
		LocalName: cfg.LocalName,

		Timeout:      time.Second * 10,
		RetryFailure: true,
	}

	sslSkipVerify, _ := strconv.ParseBool(uri.Query().Get("skip_ssl_verify"))

	serverName := uri.Query().Get("server_name")
	if serverName == "" {
		serverName = uri.Hostname()
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: sslSkipVerify, //#nosec G402 -- This is ok (and required!) because it is configurable and disabled by default.
		Certificates:       tlsCertificates,
		ServerName:         serverName,
		MinVersion:         tls.VersionTLS12,
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
			dialer.TLSConfig = tlsConfig
			// Enforcing StartTLS
			dialer.StartTLSPolicy = gomail.MandatoryStartTLS
		}
	case "smtps":
		dialer.TLSConfig = tlsConfig
		dialer.SSL = true
	}

	return &SMTPClient{
		Dialer: dialer,
	}, nil
}

func (c *courier) QueueEmail(ctx context.Context, t EmailTemplate) (uuid.UUID, error) {
	recipient, err := t.EmailRecipient()
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	if _, err := mail.ParseAddress(recipient); err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	subject, err := t.EmailSubject(ctx)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	bodyPlaintext, err := t.EmailBodyPlaintext(ctx)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	templateData, err := json.Marshal(t)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	message := &Message{
		Status:       MessageStatusQueued,
		Type:         MessageTypeEmail,
		Channel:      "email",
		Recipient:    recipient,
		Body:         bodyPlaintext,
		Subject:      subject,
		TemplateType: t.TemplateType(),
		TemplateData: templateData,
	}

	if err := c.deps.CourierPersister().AddMessage(ctx, message); err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	return message.ID, nil
}
