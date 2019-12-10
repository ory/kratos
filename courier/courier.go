package courier

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/x"
)

type (
	smtpDependencies interface {
		PersistenceProvider
		x.LoggingProvider
	}
	Courier struct {
		dialer *gomail.Dialer
		d      smtpDependencies
		c      configuration.Provider
	}
	Provider interface {
		Courier() *Courier
	}
)

func NewSMTP(d smtpDependencies, c configuration.Provider) *Courier {
	uri := c.CourierSMTPURL()
	sslSkipVerify, _ := strconv.ParseBool(uri.Query().Get("skip_ssl_verify"))
	password, _ := uri.User.Password()
	port, _ := strconv.ParseInt(uri.Port(), 10, 64)
	return &Courier{
		d: d,
		c: c,
		dialer: &gomail.Dialer{
			Host:      uri.Hostname(),
			Port:      int(port),
			Username:  uri.User.Username(),
			Password:  password,
			SSL:       uri.Scheme == "smtps",
			TLSConfig: &tls.Config{InsecureSkipVerify: sslSkipVerify},
		},
	}
}

func (m *Courier) SendEmail(ctx context.Context, t EmailTemplate) (uuid.UUID, error) {
	body, err := t.EmailBody()
	if err != nil {
		return uuid.Nil, err
	}

	subject, err := t.EmailBody()
	if err != nil {
		return uuid.Nil, err
	}

	recipient, err := t.EmailRecipient()
	if err != nil {
		return uuid.Nil, err
	}

	message := &Message{
		Status:    MessageStatusQueued,
		Type:      MessageTypeEmail,
		Body:      body,
		Subject:   subject,
		Recipient: recipient,
	}
	if err := m.d.CourierPersister().AddMessage(ctx, message); err != nil {
		return uuid.Nil, err
	}
	return message.ID, nil
}

func (m *Courier) Work(ctx context.Context) error {
	errChan := make(chan error)
	defer close(errChan)

	go m.watchMessages(ctx, errChan)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func (m *Courier) watchMessages(ctx context.Context, errChan chan error) {
	for {
		if err := backoff.Retry(func() error {
			messages, err := m.d.CourierPersister().NextMessages(ctx, 10)
			if err != nil {
				return err
			}

			for k := range messages {
				var msg Message = messages[k]

				switch msg.Type {
				case MessageTypeEmail:
					from := m.c.CourierSMTPFrom()
					gm := gomail.NewMessage()
					gm.SetHeader("From", from)
					gm.SetHeader("To", msg.Recipient)
					gm.SetHeader("Subject", msg.Subject)
					gm.SetBody("text/plain", msg.Body)
					gm.AddAlternative("text/html", msg.Body)

					if err := m.dialer.DialAndSend(gm); err != nil {
						m.d.Logger().
							WithError(err).
							WithField("smtp_server", fmt.Sprintf("smtp(s)://%s:%d", m.dialer.Host, m.dialer.Port)).
							WithField("email_to", msg.Recipient).WithField("email_from", from).
							Error("Unable to send email using SMTP connection.")
						continue
					}

					if err := m.d.CourierPersister().SetMessageStatus(ctx, msg.ID, MessageStatusSent); err != nil {
						m.d.Logger().
							WithError(err).
							WithField("message_id", msg.ID).
							Error(`Unable to set the message status to "sent".`)
						return err
					}
				default:
					return errors.Errorf("received unexpected message type: %d", msg.Type)
				}
			}

			return nil
		}, backoff.NewExponentialBackOff()); err != nil {
			errChan <- err
			return
		}
		time.Sleep(time.Second)
	}
}
