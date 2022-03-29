package courier

import (
	"context"
	"time"

	"github.com/ory/kratos/courier/template"

	"github.com/cenkalti/backoff"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
	gomail "github.com/ory/mail/v3"
)

type (
	Dependencies interface {
		PersistenceProvider
		x.LoggingProvider
		ConfigProvider
		x.HTTPClientProvider
	}

	Courier interface {
		Work(ctx context.Context) error
		QueueEmail(ctx context.Context, t EmailTemplate) (uuid.UUID, error)
		QueueSMS(ctx context.Context, t SMSTemplate) (uuid.UUID, error)
		SmtpDialer() *gomail.Dialer
		DispatchQueue(ctx context.Context) error
		DispatchMessage(ctx context.Context, msg Message) error
		SetGetEmailTemplateType(f func(t EmailTemplate) (TemplateType, error))
		SetNewEmailTemplateFromMessage(f func(d template.Dependencies, msg Message) (EmailTemplate, error))
		UseBackoff(b backoff.BackOff)
	}

	Provider interface {
		Courier(ctx context.Context) Courier
	}

	ConfigProvider interface {
		CourierConfig(ctx context.Context) config.CourierConfigs
	}

	courier struct {
		smsClient   *smsClient
		smtpClient  *smtpClient
		deps        Dependencies
		failOnError bool
		backoff     backoff.BackOff
	}
)

func NewCourier(ctx context.Context, deps Dependencies) Courier {
	return &courier{
		smsClient:  newSMS(ctx, deps),
		smtpClient: newSMTP(ctx, deps),
		deps:       deps,
		backoff:    backoff.NewExponentialBackOff(),
	}
}

func (c *courier) FailOnDispatchError() {
	c.failOnError = true
}

func (c *courier) Work(ctx context.Context) error {
	errChan := make(chan error)
	defer close(errChan)

	go c.watchMessages(ctx, errChan)

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func (c *courier) UseBackoff(b backoff.BackOff) {
	c.backoff = b
}

func (c *courier) watchMessages(ctx context.Context, errChan chan error) {
	c.backoff.Reset()
	for {
		if err := backoff.Retry(func() error {
			return c.DispatchQueue(ctx)
		}, c.backoff); err != nil {
			errChan <- err
			return
		}
		time.Sleep(time.Second)
	}
}
