// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"time"

	"github.com/ory/kratos/courier/template"
	"github.com/ory/x/jsonnetsecure"

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
		x.TracingProvider
		x.LoggingProvider
		ConfigProvider
		x.HTTPClientProvider
		jsonnetsecure.VMProvider
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
		FailOnDispatchError()
	}

	Provider interface {
		Courier(ctx context.Context) (Courier, error)
	}

	ConfigProvider interface {
		CourierConfig() config.CourierConfigs
	}

	courier struct {
		smsClient           *smsClient
		smtpClient          *smtpClient
		httpClient          *httpClient
		deps                Dependencies
		failOnDispatchError bool
		backoff             backoff.BackOff
	}
)

func NewCourier(ctx context.Context, deps Dependencies) (Courier, error) {
	smtp, err := newSMTP(ctx, deps)
	if err != nil {
		return nil, err
	}

	expBackoff := backoff.NewExponentialBackOff()
	// never stop retrying
	expBackoff.MaxElapsedTime = 0

	return &courier{
		smsClient:  newSMS(ctx, deps),
		smtpClient: smtp,
		httpClient: newHTTP(ctx, deps),
		deps:       deps,
		backoff:    expBackoff,
	}, nil
}

func (c *courier) FailOnDispatchError() {
	c.failOnDispatchError = true
}

func (c *courier) Work(ctx context.Context) error {
	wait := c.deps.CourierConfig().CourierWorkerPullWait(ctx)
	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
			return ctx.Err()
		case <-time.After(wait):
			if err := backoff.Retry(func() error {
				if err := c.DispatchQueue(ctx); err != nil {
					return err
				}
				// when we succeed, we want to reset the backoff
				c.backoff.Reset()
				return nil
			}, c.backoff); err != nil {
				return err
			}
		}
	}
}

func (c *courier) UseBackoff(b backoff.BackOff) {
	c.backoff = b
}
