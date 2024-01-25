// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/request"
	"github.com/ory/kratos/x"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/otelx"
)

type (
	httpChannel struct {
		id            string
		requestConfig json.RawMessage
		d             channelDependencies
	}
	channelDependencies interface {
		x.TracingProvider
		x.LoggingProvider
		x.HTTPClientProvider
		jsonnetsecure.VMProvider
		ConfigProvider
	}
)

var _ Channel = new(httpChannel)

func newHttpChannel(id string, requestConfig json.RawMessage, d channelDependencies) *httpChannel {
	return &httpChannel{
		id:            id,
		requestConfig: requestConfig,
		d:             d,
	}
}

func (c *httpChannel) ID() string {
	return c.id
}

type httpDataModel struct {
	Recipient    string
	Subject      string
	Body         string
	TemplateType template.TemplateType
	TemplateData Template
	MessageType  string
}

func (c *httpChannel) Dispatch(ctx context.Context, msg Message) (err error) {
	ctx, span := c.d.Tracer(ctx).Tracer().Start(ctx, "courier.httpChannel.Dispatch")
	defer otelx.End(span, &err)

	builder, err := request.NewBuilder(ctx, c.requestConfig, c.d, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	tmpl, err := newTemplate(c.d, msg)
	if err != nil {
		return errors.WithStack(err)
	}

	td := httpDataModel{
		Recipient:    msg.Recipient,
		Subject:      msg.Subject,
		Body:         msg.Body,
		TemplateType: msg.TemplateType,
		TemplateData: tmpl,
		MessageType:  msg.Type.String(),
	}

	req, err := builder.BuildRequest(ctx, td)
	if err != nil {
		return errors.WithStack(err)
	}
	req = req.WithContext(ctx)

	res, err := c.d.HTTPClient(ctx).Do(req)
	if err != nil {
		return errors.WithStack(err)
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		c.d.Logger().
			WithField("message_id", msg.ID).
			WithField("message_type", msg.Type).
			WithField("message_template_type", msg.TemplateType).
			WithField("message_subject", msg.Subject).
			Debug("Courier sent out mailer.")
		return nil
	}

	err = errors.Errorf(
		"unable to dispatch mail delivery because upstream server replied with status code %d",
		res.StatusCode,
	)
	c.d.Logger().
		WithField("message_id", msg.ID).
		WithField("message_type", msg.Type).
		WithField("message_template_type", msg.TemplateType).
		WithField("message_subject", msg.Subject).
		WithError(err).
		Error("sending mail via HTTP failed.")
	return errors.WithStack(err)
}

func newTemplate(d template.Dependencies, msg Message) (Template, error) {
	switch msg.Type {
	case MessageTypeEmail:
		return NewEmailTemplateFromMessage(d, msg)
	case MessageTypeSMS:
		return NewSMSTemplateFromMessage(d, msg)
	default:
		return nil, fmt.Errorf("received unexpected message type: %s", msg.Type)
	}
}
