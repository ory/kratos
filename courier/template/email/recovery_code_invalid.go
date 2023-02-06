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
	RecoveryCodeInvalid struct {
		deps  template.Dependencies
		model *RecoveryCodeInvalidModel
	}
	RecoveryCodeInvalidModel struct {
		To string
	}
)

func NewRecoveryCodeInvalid(d template.Dependencies, m *RecoveryCodeInvalidModel) *RecoveryCodeInvalid {
	return &RecoveryCodeInvalid{deps: d, model: m}
}

func (t *RecoveryCodeInvalid) EmailRecipient() (string, error) {
	return t.model.To, nil
}

func (t *RecoveryCodeInvalid) EmailSubject(ctx context.Context) (string, error) {
	filesystem := os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx))
	remoteURL := t.deps.CourierConfig().CourierTemplatesRecoveryCodeInvalid(ctx).Subject

	subject, err := template.LoadText(ctx, t.deps, filesystem, "recovery_code/invalid/email.subject.gotmpl", "recovery_code/invalid/email.subject*", t.model, remoteURL)

	return strings.TrimSpace(subject), err
}

func (t *RecoveryCodeInvalid) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.deps, os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)), "recovery_code/invalid/email.body.gotmpl", "recovery_code/invalid/email.body*", t.model, t.deps.CourierConfig().CourierTemplatesRecoveryCodeInvalid(ctx).Body.HTML)
}

func (t *RecoveryCodeInvalid) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.deps, os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)), "recovery_code/invalid/email.body.plaintext.gotmpl", "recovery_code/invalid/email.body.plaintext*", t.model, t.deps.CourierConfig().CourierTemplatesRecoveryCodeInvalid(ctx).Body.PlainText)
}

func (t *RecoveryCodeInvalid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.model)
}
