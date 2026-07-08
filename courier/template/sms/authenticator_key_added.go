// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package sms

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/ory/kratos/courier/template"
)

type (
	AuthenticatorKeyAdded struct {
		deps  template.Dependencies
		model *AuthenticatorKeyAddedModel
	}
	AuthenticatorKeyAddedModel struct {
		To                 string         `json:"to"`
		Identity           map[string]any `json:"identity"`
		AddedAt            string         `json:"added_at"`
		TransientPayload   map[string]any `json:"transient_payload"`
		UserRequestHeaders http.Header    `json:"-"`
	}
)

func NewAuthenticatorKeyAdded(d template.Dependencies, m *AuthenticatorKeyAddedModel) *AuthenticatorKeyAdded {
	return &AuthenticatorKeyAdded{deps: d, model: m}
}

func (t *AuthenticatorKeyAdded) PhoneNumber() (string, error) {
	return t.model.To, nil
}

func (t *AuthenticatorKeyAdded) SMSBody(ctx context.Context) (string, error) {
	return template.LoadText(
		ctx,
		t.deps,
		os.DirFS(t.deps.CourierConfig().CourierTemplatesRoot(ctx)),
		"authenticator_key_added/sms.body.gotmpl",
		"authenticator_key_added/sms.body*",
		t.model,
		t.deps.CourierConfig().CourierSMSTemplatesAuthenticatorKeyAdded(ctx).Body.PlainText,
	)
}

func (t *AuthenticatorKeyAdded) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.model)
}

func (t *AuthenticatorKeyAdded) TemplateType() template.TemplateType {
	return template.TypeAuthenticatorKeyAdded
}

func (t *AuthenticatorKeyAdded) RequestHeaders() http.Header {
	return t.model.UserRequestHeaders
}
