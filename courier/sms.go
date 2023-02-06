// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/request"
)

type sendSMSRequestBody struct {
	From string `json:"from"`
	To   string `json:"to"`
	Body string `json:"body"`
}

type smsClient struct {
	RequestConfig json.RawMessage

	GetTemplateType        func(t SMSTemplate) (TemplateType, error)
	NewTemplateFromMessage func(d Dependencies, msg Message) (SMSTemplate, error)
}

func newSMS(ctx context.Context, deps Dependencies) *smsClient {
	return &smsClient{
		RequestConfig:          deps.CourierConfig().CourierSMSRequestConfig(ctx),
		GetTemplateType:        SMSTemplateType,
		NewTemplateFromMessage: NewSMSTemplateFromMessage,
	}
}

func (c *courier) QueueSMS(ctx context.Context, t SMSTemplate) (uuid.UUID, error) {
	recipient, err := t.PhoneNumber()
	if err != nil {
		return uuid.Nil, err
	}

	templateType, err := c.smsClient.GetTemplateType(t)
	if err != nil {
		return uuid.Nil, err
	}

	templateData, err := json.Marshal(t)
	if err != nil {
		return uuid.Nil, err
	}

	message := &Message{
		Status:       MessageStatusQueued,
		Type:         MessageTypePhone,
		Recipient:    recipient,
		TemplateType: templateType,
		TemplateData: templateData,
	}
	if err := c.deps.CourierPersister().AddMessage(ctx, message); err != nil {
		return uuid.Nil, err
	}

	return message.ID, nil
}

func (c *courier) dispatchSMS(ctx context.Context, msg Message) error {
	if !c.deps.CourierConfig().CourierSMSEnabled(ctx) {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Courier tried to deliver an sms but courier.sms.enabled is set to false!"))
	}

	tmpl, err := c.smsClient.NewTemplateFromMessage(c.deps, msg)
	if err != nil {
		return err
	}

	body, err := tmpl.SMSBody(ctx)
	if err != nil {
		return err
	}

	builder, err := request.NewBuilder(c.smsClient.RequestConfig, c.deps)
	if err != nil {
		return err
	}

	req, err := builder.BuildRequest(ctx, &sendSMSRequestBody{
		To:   msg.Recipient,
		From: c.deps.CourierConfig().CourierSMSFrom(ctx),
		Body: body,
	})
	if err != nil {
		return err
	}

	res, err := c.deps.HTTPClient(ctx).Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	default:
		return errors.New(http.StatusText(res.StatusCode))
	}

	return nil
}
