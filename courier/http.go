// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ory/kratos/request"
	"github.com/ory/x/otelx"
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
func (c *courier) dispatchMailerEmail(ctx context.Context, msg Message) (err error) {
	ctx, span := c.deps.Tracer(ctx).Tracer().Start(ctx, "courier.http.dispatchMailerEmail")
	defer otelx.End(span, &err)

	builder, err := request.NewBuilder(ctx, c.httpClient.RequestConfig, c.deps)
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

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		c.deps.Logger().
			WithField("message_id", msg.ID).
			WithField("message_type", msg.Type).
			WithField("message_template_type", msg.TemplateType).
			WithField("message_subject", msg.Subject).
			Debug("Courier sent out mailer.")
		return nil
	}

	err = fmt.Errorf(
		"unable to dispatch mail delivery because upstream server replied with status code %d",
		res.StatusCode,
	)
	c.deps.Logger().
		WithField("message_id", msg.ID).
		WithField("message_type", msg.Type).
		WithField("message_template_type", msg.TemplateType).
		WithField("message_subject", msg.Subject).
		WithError(err).
		Error("sending mail via HTTP failed.")
	return err
}
