// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"time"

	"github.com/ory/x/jsonnetsecure"

	"github.com/cenkalti/backoff"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
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
		DispatchQueue(ctx context.Context) error
		DispatchMessage(ctx context.Context, msg Message) error
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
		courierChannels     map[string]Channel
		deps                Dependencies
		failOnDispatchError bool
		backoff             backoff.BackOff
	}
)

func NewCourier(ctx context.Context, deps Dependencies) (Courier, error) {
	return NewCourierWithCustomTemplates(ctx, deps, NewEmailTemplateFromMessage)
}

func NewCourierWithCustomTemplates(ctx context.Context, deps Dependencies, newEmailTemplateFromMessage func(d template.Dependencies, msg Message) (EmailTemplate, error)) (Courier, error) {
	cs, err := deps.CourierConfig().CourierChannels(ctx)
	if err != nil {
		return nil, err
	}
	channels := make(map[string]Channel, len(cs))
	for _, c := range cs {
		switch c.Type {
		case "smtp":
			ch, err := NewSMTPChannelWithCustomTemplates(deps, c.SMTPConfig, newEmailTemplateFromMessage)
			if err != nil {
				return nil, err
			}
			channels[ch.ID()] = ch
		case "http":
			channels[c.ID] = newHttpChannel(c.ID, c.RequestConfig, deps)
		default:
			return nil, errors.Errorf("unknown courier channel type: %s", c.Type)
		}
	}

	return &courier{
		deps:            deps,
		backoff:         backoff.NewExponentialBackOff(),
		courierChannels: channels,
	}, nil
}

func (c *courier) FailOnDispatchError() {
	c.failOnDispatchError = true
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
	wait := c.deps.CourierConfig().CourierWorkerPullWait(ctx)
	c.backoff.Reset()
	for {
		if err := backoff.Retry(func() error {
			return c.DispatchQueue(ctx)
		}, c.backoff); err != nil {
			errChan <- err
			return
		}
		time.Sleep(wait)
	}
}
