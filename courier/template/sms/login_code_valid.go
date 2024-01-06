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
	LoginCodeValid struct {
		deps  template.Dependencies
		model *LoginCodeValidModel
	}
	LoginCodeValidModel struct {
		To        string
		LoginCode string
		Identity  map[string]interface{}
	}
)

func NewLoginCodeValid(d template.Dependencies, m *LoginCodeValidModel) *LoginCodeValid {
	return &LoginCodeValid{deps: d, model: m}
}

func (t *LoginCodeValid) PhoneNumber() (string, error) {
	return t.model.To, nil
}

func (t *LoginCodeValid) SMSBody(ctx context.Context) (string, error) {
	return template.LoadText(
		ctx,
		t.deps,
		os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)),
		"login_code/valid/sms.body.gotmpl", // TODO:!!!
		"login_code/valid/sms.body*",
		t.model,
		"",
	)
}

func (t *LoginCodeValid) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.model)
}

func (t *LoginCodeValid) TemplateType() template.TemplateType {
	return template.TypeLoginCodeValid
}
