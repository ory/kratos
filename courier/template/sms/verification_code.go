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
	VerificationCodeValid struct {
		deps  template.Dependencies
		model *VerificationCodeValidModel
	}

	VerificationCodeValidModel struct {
		To               string                 `json:"to"`
		VerificationCode string                 `json:"verification_code"`
		Identity         map[string]interface{} `json:"identity"`
		RequestURL       string                 `json:"request_url"`
		TransientPayload map[string]interface{} `json:"transient_payload"`
		ExpiresInMinutes int                    `json:"expires_in_minutes"`
	}
)

func NewVerificationCodeValid(d template.Dependencies, m *VerificationCodeValidModel) *VerificationCodeValid {
	return &VerificationCodeValid{deps: d, model: m}
}

func (t *VerificationCodeValid) PhoneNumber() (string, error) {
	return t.model.To, nil
}

func (t *VerificationCodeValid) SMSBody(ctx context.Context) (string, error) {
	return template.LoadText(
		ctx,
		t.deps,
		os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)),
		"verification_code/valid/sms.body.gotmpl",
		"verification_code/valid/sms.body*",
		t.model,
		t.deps.CourierConfig().CourierSMSTemplatesVerificationCodeValid(ctx).Body.PlainText,
	)
}

func (t *VerificationCodeValid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.model)
}

func (t *VerificationCodeValid) TemplateType() template.TemplateType {
	return template.TypeVerificationCodeValid
}
