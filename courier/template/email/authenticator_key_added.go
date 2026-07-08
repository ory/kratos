// Copyright © 2026 Ory Corp
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
	AuthenticatorKeyAdded struct {
		d template.Dependencies
		m *AuthenticatorKeyAddedModel
	}
	AuthenticatorKeyAddedModel struct {
		To               string         `json:"to"`
		Identity         map[string]any `json:"identity"`
		AddedAt          string         `json:"added_at"`
		TransientPayload map[string]any `json:"transient_payload"`
	}
)

func NewAuthenticatorKeyAdded(d template.Dependencies, m *AuthenticatorKeyAddedModel) *AuthenticatorKeyAdded {
	return &AuthenticatorKeyAdded{d: d, m: m}
}

func (t *AuthenticatorKeyAdded) EmailRecipient() (string, error) {
	return t.m.To, nil
}

func (t *AuthenticatorKeyAdded) EmailSubject(ctx context.Context) (string, error) {
	subject, err := template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "authenticator_key_added/email.subject.gotmpl", "authenticator_key_added/email.subject*", t.m, t.d.CourierConfig().CourierTemplatesAuthenticatorKeyAdded(ctx).Subject)
	return strings.TrimSpace(subject), err
}

func (t *AuthenticatorKeyAdded) EmailBody(ctx context.Context) (string, error) {
	return template.LoadHTML(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "authenticator_key_added/email.body.gotmpl", "authenticator_key_added/email.body*", t.m, t.d.CourierConfig().CourierTemplatesAuthenticatorKeyAdded(ctx).Body.HTML)
}

func (t *AuthenticatorKeyAdded) EmailBodyPlaintext(ctx context.Context) (string, error) {
	return template.LoadText(ctx, t.d, os.DirFS(t.d.CourierConfig().CourierTemplatesRoot(ctx)), "authenticator_key_added/email.body.plaintext.gotmpl", "authenticator_key_added/email.body.plaintext*", t.m, t.d.CourierConfig().CourierTemplatesAuthenticatorKeyAdded(ctx).Body.PlainText)
}

func (t *AuthenticatorKeyAdded) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.m)
}

func (t *AuthenticatorKeyAdded) TemplateType() template.TemplateType {
	return template.TypeAuthenticatorKeyAdded
}
