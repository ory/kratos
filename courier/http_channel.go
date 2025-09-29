// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"fmt"
	"io"

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
		requestConfig *request.Config
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

func newHttpChannel(id string, requestConfig *request.Config, d channelDependencies) *httpChannel {
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
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	// Optional HTMLBody contains the HTML version of an email template when available.
	HTMLBody     string                `json:"html_body,omitempty"`
	TemplateType template.TemplateType `json:"template_type"`
	TemplateData Template              `json:"template_data"`
	MessageType  string                `json:"message_type"`
}

func (c *httpChannel) Dispatch(ctx context.Context, msg Message) (err error) {
	ctx, span := c.d.Tracer(ctx).Tracer().Start(ctx, "courier.httpChannel.Dispatch")
	defer otelx.End(span, &err)

	builder, err := request.NewBuilder(ctx, c.requestConfig, c.d)
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

	c.tryPopulateHTMLBody(ctx, tmpl, &td)

	req, err := builder.BuildRequest(ctx, td)
	if err != nil {
		return errors.WithStack(err)
	}
	req = req.WithContext(ctx)

	res, err := c.d.HTTPClient(ctx).Do(req)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() { _ = res.Body.Close() }()
	res.Body = io.NopCloser(io.LimitReader(res.Body, 1024))

	logger := c.d.Logger().
		WithField("http_server", c.requestConfig.URL).
		WithField("message_id", msg.ID).
		WithField("message_nid", msg.NID).
		WithField("message_type", msg.Type).
		WithField("message_template_type", msg.TemplateType).
		WithField("message_subject", msg.Subject)

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		logger.Debug("Courier sent out mailer.")
		return nil
	}

	err = errors.Errorf(
		"unable to dispatch mail delivery because upstream server replied with status code %d",
		res.StatusCode,
	)

	body, _ := io.ReadAll(res.Body)
	logger.
		WithError(err).
		WithField("http_response_body", string(body)).
		Error("sending mail via HTTP failed.")

	return errors.WithStack(err)
}

func (c *httpChannel) tryPopulateHTMLBody(ctx context.Context, tmpl Template, td *httpDataModel) {
	if emailTmpl, ok := tmpl.(EmailTemplate); ok {
		// Only get the HTML body from the template; plaintext body comes from msg.Body
		// to maintain backward compatibility with existing behavior
		if htmlBody, err := emailTmpl.EmailBody(ctx); err != nil {
			c.d.Logger().WithError(err).Error("Unable to get email HTML body from template.")
		} else {
			td.HTMLBody = htmlBody
		}
	}
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
