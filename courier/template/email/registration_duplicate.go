// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package email

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/ory/kratos/courier/template"
)

type (
	RegistrationDuplicate struct {
		d template.Dependencies
		m *RegistrationDuplicateModel
	}
	RegistrationDuplicateModel struct {
		To               string                 `json:"to"`
		Identity         map[string]interface{} `json:"identity"`
		RequestURL       string                 `json:"request_url"`
		TransientPayload map[string]interface{} `json:"transient_payload"`
	}
)

func NewRegistrationDuplicate(d template.Dependencies, m *RegistrationDuplicateModel) *RegistrationDuplicate {
	return &RegistrationDuplicate{d: d, m: m}
}

func (t *RegistrationDuplicate) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *RegistrationDuplicate) EmailSubject(ctx context.Context) (string, error) {
	subject, err := template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "registration/duplicate/email.subject.gotmpl", "registration/duplicate/email.subject*", t.m, t.d.CourierConfig().CourierTemplatesRegistrationDuplicate(ctx).Subject)

	return strings.TrimSpace(subject), err
}

func (t *RegistrationDuplicate) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "registration/duplicate/email.body.gotmpl", "registration/duplicate/email.body*", t.m, t.d.CourierConfig().CourierTemplatesRegistrationDuplicate(ctx).Body.HTML)
}

func (t *RegistrationDuplicate) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "registration/duplicate/email.body.plaintext.gotmpl", "registration/duplicate/email.body.plaintext*", t.m, t.d.CourierConfig().CourierTemplatesRegistrationDuplicate(ctx).Body.PlainText)
}

func (t *RegistrationDuplicate) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}

func (t *RegistrationDuplicate) TemplateType() template.TemplateType {
	return template.TypeRegistrationDuplicateEmail
}
