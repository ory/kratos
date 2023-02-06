// Copyright © 2023 Ory Corp
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
	VerificationValid struct {
		d template.Dependencies
		m *VerificationValidModel
	}
	VerificationValidModel struct {
		To              string
		VerificationURL string
		Identity        map[string]interface{}
	}
)

func NewVerificationValid(d template.Dependencies, m *VerificationValidModel) *VerificationValid {
	return &VerificationValid{d: d, m: m}
}

func (t *VerificationValid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *VerificationValid) EmailSubject(ctx context.Context) (string, error) {
	subject, err := template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "verification/valid/email.subject.gotmpl", "verification/valid/email.subject*", t.m, t.d.CourierConfig().CourierTemplatesVerificationValid(ctx).Subject)

	return strings.TrimSpace(subject), err
}

func (t *VerificationValid) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "verification/valid/email.body.gotmpl", "verification/valid/email.body*", t.m, t.d.CourierConfig().CourierTemplatesVerificationValid(ctx).Body.HTML)
}

func (t *VerificationValid) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "verification/valid/email.body.plaintext.gotmpl", "verification/valid/email.body.plaintext*", t.m, t.d.CourierConfig().CourierTemplatesVerificationValid(ctx).Body.PlainText)
}

func (t *VerificationValid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
