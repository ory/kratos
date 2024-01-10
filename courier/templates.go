// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package courier

import (
	"context"
	"encoding/json"

	"github.com/ory/kratos/courier/template"

	"github.com/pkg/errors"

	"github.com/ory/kratos/courier/template/email"
)

type (
	Template interface {
		json.Marshaler
		TemplateType() template.TemplateType
	}

	EmailTemplate interface {
		Template
		EmailSubject(context.Context) (string, error)
		EmailBody(context.Context) (string, error)
		EmailBodyPlaintext(context.Context) (string, error)
		EmailRecipient() (string, error)
	}
)

func NewEmailTemplateFromMessage(d template.Dependencies, msg Message) (EmailTemplate, error) {
	switch msg.TemplateType {
	case template.TypeRecoveryInvalid:
		var t email.RecoveryInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewRecoveryInvalid(d, &t), nil
	case template.TypeRecoveryValid:
		var t email.RecoveryValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewRecoveryValid(d, &t), nil
	case template.TypeRecoveryCodeInvalid:
		var t email.RecoveryCodeInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewRecoveryCodeInvalid(d, &t), nil
	case template.TypeRecoveryCodeValid:
		var t email.RecoveryCodeValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewRecoveryCodeValid(d, &t), nil
	case template.TypeVerificationInvalid:
		var t email.VerificationInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewVerificationInvalid(d, &t), nil
	case template.TypeVerificationValid:
		var t email.VerificationValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewVerificationValid(d, &t), nil
	case template.TypeVerificationCodeInvalid:
		var t email.VerificationCodeInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewVerificationCodeInvalid(d, &t), nil
	case template.TypeVerificationCodeValid:
		var t email.VerificationCodeValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewVerificationCodeValid(d, &t), nil
	case template.TypeTestStub:
		var t email.TestStubModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewTestStub(d, &t), nil
	case template.TypeLoginCodeValid:
		var t email.LoginCodeValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewLoginCodeValid(d, &t), nil
	case template.TypeRegistrationCodeValid:
		var t email.RegistrationCodeValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewRegistrationCodeValid(d, &t), nil
	default:
		return nil, errors.Errorf("received unexpected message template type: %s", msg.TemplateType)
	}
}
