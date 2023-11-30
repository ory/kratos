// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/mail"
	"strconv"
	"time"

	"github.com/gofrs/uuid"

	gomail "github.com/ory/mail/v3"
)

type SMTPClient struct {
	*gomail.Dialer
}

func NewSMTP(ctx context.Context, deps Dependencies) (*SMTPClient, error) {
	uri, err := deps.CourierConfig().CourierSMTPURL(ctx)
	if err != nil {
		return nil, err
	}

	var tlsCertificates []tls.Certificate
	clientCertPath := deps.CourierConfig().CourierSMTPClientCertPath(ctx)
	clientKeyPath := deps.CourierConfig().CourierSMTPClientKeyPath(ctx)

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

	localName := deps.CourierConfig().CourierSMTPLocalName(ctx)
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
			//#nosec G402 -- This is ok (and required!) because it is configurable and disabled by default.
			dialer.TLSConfig = &tls.Config{InsecureSkipVerify: sslSkipVerify, Certificates: tlsCertificates, ServerName: serverName}
			// Enforcing StartTLS
			dialer.StartTLSPolicy = gomail.MandatoryStartTLS
		}
	case "smtps":
		//#nosec G402 -- This is ok (and required!) because it is configurable and disabled by default.
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: sslSkipVerify, Certificates: tlsCertificates, ServerName: serverName}
		dialer.SSL = true
	}

	return &SMTPClient{
		Dialer: dialer,
	}, nil
}

func (c *courier) QueueEmail(ctx context.Context, t EmailTemplate) (uuid.UUID, error) {
	recipient, err := t.EmailRecipient()
	if err != nil {
		return uuid.Nil, err
	}
	if _, err := mail.ParseAddress(recipient); err != nil {
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
		TemplateType: t.TemplateType(),
		TemplateData: templateData,
	}

	if err := c.deps.CourierPersister().AddMessage(ctx, message); err != nil {
		return uuid.Nil, err
	}

	return message.ID, nil
}
