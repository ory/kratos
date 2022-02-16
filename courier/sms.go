package courier

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/request"
)

type sendSMSRequestBody struct {
	To   string
	From string
	Body string
}

type smsClient struct {
	*http.Client
	RequestConfig json.RawMessage

	GetTemplateType        func(t SMSTemplate) (TemplateType, error)
	NewTemplateFromMessage func(d Dependencies, msg Message) (SMSTemplate, error)
}

func newSMS(ctx context.Context, deps Dependencies) *smsClient {
	if !deps.CourierConfig(ctx).CourierSMSEnabled() {
		deps.Logger().Error("messages will not be sent - no sms gate server address is set in config")
	}

	return &smsClient{
		Client:        &http.Client{},
		RequestConfig: deps.CourierConfig(ctx).CourierSMSRequestConfig(),

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
	tmpl, err := c.smsClient.NewTemplateFromMessage(c.deps, msg)
	if err != nil {
		return err
	}

	body, err := tmpl.SMSBody(ctx)
	if err != nil {
		return err
	}

	builder, err := request.NewBuilder(c.smsClient.RequestConfig, c.deps.HTTPClient(ctx), c.deps.Logger())
	if err != nil {
		return err
	}

	req, err := builder.BuildRequest(&sendSMSRequestBody{
		To:   msg.Recipient,
		From: c.deps.CourierConfig(ctx).CourierSMSFrom(),
		Body: body,
	})
	if err != nil {
		return err
	}

	res, err := c.smsClient.Do(req)
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
