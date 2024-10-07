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
	RegistrationCodeValid struct {
		deps  template.Dependencies
		model *RegistrationCodeValidModel
	}
	RegistrationCodeValidModel struct {
		To               string                 `json:"to"`
		RegistrationCode string                 `json:"registration_code"`
		Identity         map[string]interface{} `json:"identity"`
		RequestURL       string                 `json:"request_url"`
		TransientPayload map[string]interface{} `json:"transient_payload"`
		ExpiresInMinutes int                    `json:"expires_in_minutes"`
	}
)

func NewRegistrationCodeValid(d template.Dependencies, m *RegistrationCodeValidModel) *RegistrationCodeValid {
	return &RegistrationCodeValid{deps: d, model: m}
}

func (t *RegistrationCodeValid) PhoneNumber() (string, error) {
	return t.model.To, nil
}

func (t *RegistrationCodeValid) SMSBody(ctx context.Context) (string, error) {
	return template.LoadText(
		ctx,
		t.deps,
		os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)),
		"registration_code/valid/sms.body.gotmpl",
		"registration_code/valid/sms.body*",
		t.model,
		t.deps.CourierConfig().CourierSMSTemplatesRegistrationCodeValid(ctx).Body.PlainText,
	)
}

func (t *RegistrationCodeValid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.model)
}

func (t *RegistrationCodeValid) TemplateType() template.TemplateType {
	return template.TypeRegistrationCodeValid
}
