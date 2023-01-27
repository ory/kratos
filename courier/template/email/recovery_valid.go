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
	RecoveryValid struct {
		d template.Dependencies
		m *RecoveryValidModel
	}
	RecoveryValidModel struct {
		To          string
		RecoveryURL string
		Identity    map[string]interface{}
	}
)

func NewRecoveryValid(d template.Dependencies, m *RecoveryValidModel) *RecoveryValid {
	return &RecoveryValid{d: d, m: m}
}

func (t *RecoveryValid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *RecoveryValid) EmailSubject(ctx context.Context) (string, error) {
	subject, err := template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "recovery/valid/email.subject.gotmpl", "recovery/valid/email.subject*", t.m, t.d.CourierConfig().CourierTemplatesRecoveryValid(ctx).Subject)

	return strings.TrimSpace(subject), err
}

func (t *RecoveryValid) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "recovery/valid/email.body.gotmpl", "recovery/valid/email.body*", t.m, t.d.CourierConfig().CourierTemplatesRecoveryValid(ctx).Body.HTML)
}

func (t *RecoveryValid) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "recovery/valid/email.body.plaintext.gotmpl", "recovery/valid/email.body.plaintext*", t.m, t.d.CourierConfig().CourierTemplatesRecoveryValid(ctx).Body.PlainText)
}

func (t *RecoveryValid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
