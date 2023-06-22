// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/kratos/request"
)

type httpDataModel struct {
	Recipient    string
	Subject      string
	Body         string
	TemplateType TemplateType
	TemplateData EmailTemplate
}

type httpClient struct {
	RequestConfig json.RawMessage
}

func newHTTP(ctx context.Context, deps Dependencies) *httpClient {
	return &httpClient{
		RequestConfig: deps.CourierConfig().CourierEmailRequestConfig(ctx),
	}
}
func (c *courier) dispatchMailerEmail(ctx context.Context, msg Message) error {
	builder, err := request.NewBuilder(c.httpClient.RequestConfig, c.deps)
	if err != nil {
		return err
	}

	tmpl, err := c.smtpClient.NewTemplateFromMessage(c.deps, msg)
	if err != nil {
		return err
	}

	td := httpDataModel{
		Recipient:    msg.Recipient,
		Subject:      msg.Subject,
		Body:         msg.Body,
		TemplateType: msg.TemplateType,
		TemplateData: tmpl,
	}

	req, err := builder.BuildRequest(ctx, td)
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

	c.deps.Logger().
		WithField("message_id", msg.ID).
		WithField("message_type", msg.Type).
		WithField("message_template_type", msg.TemplateType).
		WithField("message_subject", msg.Subject).
		Debug("Courier sent out mailer.")

	return nil
}
