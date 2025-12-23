// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package email

import (
	"context"
	"encoding/json"

	"github.com/ory/kratos/courier/template"
)

type (
	TestStub struct {
		m *TestStubModel
	}
	TestStubModel struct {
		To       string `json:"to"`
		Subject  string `json:"subject"`
		Body     string `json:"body"`
		HTMLBody string `json:"html_body,omitempty"`
	}
)

func NewTestStub(m *TestStubModel) *TestStub {
	return &TestStub{m: m}
}

func (t *TestStub) EmailRecipient() (string, error)              { return t.m.To, nil }
func (t *TestStub) EmailSubject(context.Context) (string, error) { return t.m.Subject, nil }

func (t *TestStub) EmailBody(ctx context.Context) (string, error) {
	if t.m.HTMLBody != "" {
		return t.m.HTMLBody, nil
	}
	return t.EmailBodyPlaintext(ctx)
}

func (t *TestStub) EmailBodyPlaintext(context.Context) (string, error) { return t.m.Body, nil }
func (t *TestStub) MarshalJSON() ([]byte, error)                       { return json.Marshal(t.m) }
func (t *TestStub) TemplateType() template.TemplateType                { return template.TypeTestStub }
