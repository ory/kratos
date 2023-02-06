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
	RecoveryInvalid struct {
		d template.Dependencies
		m *RecoveryInvalidModel
	}
	RecoveryInvalidModel struct {
		To string
	}
)

func NewRecoveryInvalid(d template.Dependencies, m *RecoveryInvalidModel) *RecoveryInvalid {
	return &RecoveryInvalid{d: d, m: m}
}

func (t *RecoveryInvalid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *RecoveryInvalid) EmailSubject(ctx context.Context) (string, error) {
	subject, err := template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "recovery/invalid/email.subject.gotmpl", "recovery/invalid/email.subject*", t.m, t.d.CourierConfig().CourierTemplatesRecoveryInvalid(ctx).Subject)

	return strings.TrimSpace(subject), err
}

func (t *RecoveryInvalid) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "recovery/invalid/email.body.gotmpl", "recovery/invalid/email.body*", t.m, t.d.CourierConfig().CourierTemplatesRecoveryInvalid(ctx).Body.HTML)
}

func (t *RecoveryInvalid) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "recovery/invalid/email.body.plaintext.gotmpl", "recovery/invalid/email.body.plaintext*", t.m, t.d.CourierConfig().CourierTemplatesRecoveryInvalid(ctx).Body.PlainText)
}

func (t *RecoveryInvalid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
