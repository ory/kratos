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
	RecoveryCodeValid struct {
		deps  template.Dependencies
		model *RecoveryCodeValidModel
	}
	RecoveryCodeValidModel struct {
		To           string
		RecoveryCode string
		Identity     map[string]interface{}
	}
)

func NewRecoveryCodeValid(d template.Dependencies, m *RecoveryCodeValidModel) *RecoveryCodeValid {
	return &RecoveryCodeValid{deps: d, model: m}
}

func (t *RecoveryCodeValid) EmailRecipient() (string, error) {
	return t.model.To, nil
}

func (t *RecoveryCodeValid) EmailSubject(ctx context.Context) (string, error) {
	subject, err := template.LoadText(ctx, t.deps, os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)), "recovery_code/valid/email.subject.gotmpl", "recovery_code/valid/email.subject*", t.model, t.deps.CourierConfig().CourierTemplatesRecoveryCodeValid(ctx).Subject)

	return strings.TrimSpace(subject), err
}

func (t *RecoveryCodeValid) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.deps, os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)), "recovery_code/valid/email.body.gotmpl", "recovery_code/valid/email.body*", t.model, t.deps.CourierConfig().CourierTemplatesRecoveryCodeValid(ctx).Body.HTML)
}

func (t *RecoveryCodeValid) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.deps, os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)), "recovery_code/valid/email.body.plaintext.gotmpl", "recovery_code/valid/email.body.plaintext*", t.model, t.deps.CourierConfig().CourierTemplatesRecoveryCodeValid(ctx).Body.PlainText)
}

func (t *RecoveryCodeValid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.model)
}
