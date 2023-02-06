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
	TestStub struct {
		d template.Dependencies
		m *TestStubModel
	}
	TestStubModel struct {
		To      string
		Subject string
		Body    string
	}
)

func NewTestStub(d template.Dependencies, m *TestStubModel) *TestStub {
	return &TestStub{d: d, m: m}
}

func (t *TestStub) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *TestStub) EmailSubject(ctx context.Context) (string, error) {
	subject, err := template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "test_stub/email.subject.gotmpl", "test_stub/email.subject*", t.m, "")

	return strings.TrimSpace(subject), err
}

func (t *TestStub) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "test_stub/email.body.gotmpl", "test_stub/email.body*", t.m, "")
}

func (t *TestStub) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "test_stub/email.body.plaintext.gotmpl", "test_stub/email.body.plaintext*", t.m, "")
}

func (t *TestStub) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}
