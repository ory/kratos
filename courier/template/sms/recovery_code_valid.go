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
	RecoveryCodeValid struct {
		deps  template.Dependencies
		model *RecoveryCodeValidModel
	}
	RecoveryCodeValidModel struct {
		To               string                 `json:"to"`
		RecoveryCode     string                 `json:"recovery_code"`
		Identity         map[string]interface{} `json:"identity"`
		RequestURL       string                 `json:"request_url"`
		TransientPayload map[string]interface{} `json:"transient_payload"`
		ExpiresInMinutes int                    `json:"expires_in_minutes"`
	}
)

func NewRecoveryCodeValid(d template.Dependencies, m *RecoveryCodeValidModel) *RecoveryCodeValid {
	return &RecoveryCodeValid{deps: d, model: m}
}

func (t *RecoveryCodeValid) PhoneNumber() (string, error) {
	return t.model.To, nil
}

func (t *RecoveryCodeValid) SMSBody(ctx context.Context) (string, error) {
	return template.LoadText(
		ctx,
		t.deps,
		os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)),
		"login_code/valid/sms.body.gotmpl",
		"login_code/valid/sms.body*",
		t.model,
		t.deps.CourierConfig().CourierSMSTemplatesRecoveryCodeValid(ctx).Body.PlainText,
	)
}

func (t *RecoveryCodeValid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.model)
}

func (t *RecoveryCodeValid) TemplateType() template.TemplateType {
	return template.TypeRecoveryCodeValid
}
