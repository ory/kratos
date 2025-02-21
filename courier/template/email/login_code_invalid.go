// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package email

import (
	"context"
	"encoding/json"
	"github.com/ory/kratos/courier/template"
	"os"
	"strings"
)

type (
	LoginCodeInvalid struct {
		d template.Dependencies
		m *LoginCodeInvalidModel
	}
	LoginCodeInvalidModel struct {
		To               string                 `json:"to"`
		RequestURL       string                 `json:"request_url"`
		TransientPayload map[string]interface{} `json:"transient_payload"`
	}
)

func NewLoginCodeInvalid(d template.Dependencies, m *LoginCodeInvalidModel) *LoginCodeInvalid {
	return &LoginCodeInvalid{d: d, m: m}
}

func (t *LoginCodeInvalid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *LoginCodeInvalid) EmailSubject(ctx context.Context) (string, error) {
	subject, err := template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "login_code/invalid/email.subject.gotmpl", "login_code/invalid/email.subject*", t.m, t.d.CourierConfig().CourierTemplatesLoginCodeInvalid(ctx).Subject)

	return strings.TrimSpace(subject), err
}

func (t *LoginCodeInvalid) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "login_code/invalid/email.body.gotmpl", "login_code/invalid/email.body*", t.m, t.d.CourierConfig().CourierTemplatesLoginCodeInvalid(ctx).Body.HTML)
}

func (t *LoginCodeInvalid) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "login_code/invalid/email.body.plaintext.gotmpl", "login_code/invalid/email.body.plaintext*", t.m, t.d.CourierConfig().CourierTemplatesLoginCodeInvalid(ctx).Body.PlainText)
}

func (t *LoginCodeInvalid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}

func (t *LoginCodeInvalid) TemplateType() template.TemplateType {
	return template.TypeLoginCodeInvalid
}
