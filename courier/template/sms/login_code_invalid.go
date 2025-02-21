// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sms

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ory/kratos/courier/template"
)

type (
	LoginCodeInvalid struct {
		deps  template.Dependencies
		model *LoginCodeInvalidModel
	}
	LoginCodeInvalidModel struct {
		To               string                 `json:"to"`
		RequestURL       string                 `json:"request_url"`
		TransientPayload map[string]interface{} `json:"transient_payload"`
	}
)

func NewLoginCodeInvalid(d template.Dependencies, m *LoginCodeInvalidModel) *LoginCodeInvalid {
	return &LoginCodeInvalid{deps: d, model: m}
}

func (t *LoginCodeInvalid) PhoneNumber() (string, error) {
	return t.model.To, nil
}

func (t *LoginCodeInvalid) SMSBody(ctx context.Context) (string, error) {
	return template.LoadText(
		ctx,
		t.deps,
		os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)),
		"login_code/invalid/sms.body.gotmpl",
		"login_code/invalid/sms.body*",
		t.model,
		t.deps.CourierConfig().CourierSMSTemplatesLoginCodeInvalid(ctx).Body.PlainText,
	)
}

func (t *LoginCodeInvalid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.model)
}

func (t *LoginCodeInvalid) TemplateType() template.TemplateType {
	return template.TypeLoginCodeInvalid
}
