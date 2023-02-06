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
	VerificationCodeInvalid struct {
		d template.Dependencies
		m *VerificationCodeInvalidModel
	}
	VerificationCodeInvalidModel struct {
		To string
	}
)

func NewVerificationCodeInvalid(d template.Dependencies, m *VerificationCodeInvalidModel) *VerificationCodeInvalid {
	return &VerificationCodeInvalid{d: d, m: m}
}

func (t *VerificationCodeInvalid) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *VerificationCodeInvalid) EmailSubject(ctx context.Context) (string, error) {
	subject, err := template.LoadText(
		ctx,
		t.d,
		os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)),
		"verification_code/invalid/email.subject.gotmpl",
		"verification_code/invalid/email.subject*",
		t.m,
		t.d.CourierConfig().CourierTemplatesVerificationCodeInvalid(ctx).Subject,
	)

	return strings.TrimSpace(subject), err
}

func (t *VerificationCodeInvalid) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(
		ctx,
		t.d,
		os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)),
		"verification_code/invalid/email.body.gotmpl",
		"verification_code/invalid/email.body*",
		t.m,
		t.d.CourierConfig().CourierTemplatesVerificationCodeInvalid(ctx).Body.HTML,
	)
}

func (t *VerificationCodeInvalid) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(
		ctx,
		t.d,
		os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)),
		"verification_code/invalid/email.body.plaintext.gotmpl",
		"verification_code/invalid/email.body.plaintext*",
		t.m,
		t.d.CourierConfig().CourierTemplatesVerificationCodeInvalid(ctx).Body.PlainText,
	)
}

func (t *VerificationCodeInvalid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
