// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sms

import (
	"context"
	"encoding/json"

	"github.com/ory/kratos/courier/template"
)

type (
	TestStub struct{ m *TestStubModel }

	TestStubModel struct {
		To       string                 `json:"to"`
		Body     string                 `json:"body"`
		Identity map[string]interface{} `json:"identity"`
	}
)

func NewTestStub(m *TestStubModel) *TestStub                { return &TestStub{m: m} }
func (t *TestStub) PhoneNumber() (string, error)            { return t.m.To, nil }
func (t *TestStub) SMSBody(context.Context) (string, error) { return t.m.Body, nil }
func (t *TestStub) MarshalJSON() ([]byte, error)            { return json.Marshal(t.m) }
func (t *TestStub) TemplateType() template.TemplateType     { return template.TypeTestStub }
