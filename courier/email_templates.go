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
	EmailTemplate interface {
		json.Marshaler
		EmailSubject(context.Context) (string, error)
		EmailBody(context.Context) (string, error)
		EmailBodyPlaintext(context.Context) (string, error)
		EmailRecipient() (string, error)
	}
)

// A Template's type
//
// swagger:enum TemplateType
type TemplateType string

const (
	TypeRecoveryInvalid         TemplateType = "recovery_invalid"
	TypeRecoveryValid           TemplateType = "recovery_valid"
	TypeRecoveryCodeInvalid     TemplateType = "recovery_code_invalid"
	TypeRecoveryCodeValid       TemplateType = "recovery_code_valid"
	TypeVerificationInvalid     TemplateType = "verification_invalid"
	TypeVerificationValid       TemplateType = "verification_valid"
	TypeVerificationCodeInvalid TemplateType = "verification_code_invalid"
	TypeVerificationCodeValid   TemplateType = "verification_code_valid"
	TypeOTP                     TemplateType = "otp"
	TypeTestStub                TemplateType = "stub"
	TypeLoginCodeValid          TemplateType = "login_code_valid"
	TypeRegistrationCodeValid   TemplateType = "registration_code_valid"
)

func GetEmailTemplateType(t EmailTemplate) (TemplateType, error) {
	switch t.(type) {
	case *email.RecoveryInvalid:
		return TypeRecoveryInvalid, nil
	case *email.RecoveryValid:
		return TypeRecoveryValid, nil
	case *email.RecoveryCodeInvalid:
		return TypeRecoveryCodeInvalid, nil
	case *email.RecoveryCodeValid:
		return TypeRecoveryCodeValid, nil
	case *email.VerificationInvalid:
		return TypeVerificationInvalid, nil
	case *email.VerificationValid:
		return TypeVerificationValid, nil
	case *email.VerificationCodeInvalid:
		return TypeVerificationCodeInvalid, nil
	case *email.VerificationCodeValid:
		return TypeVerificationCodeValid, nil
	case *email.LoginCodeValid:
		return TypeLoginCodeValid, nil
	case *email.RegistrationCodeValid:
		return TypeRegistrationCodeValid, nil
	case *email.TestStub:
		return TypeTestStub, nil
	default:
		return "", errors.Errorf("unexpected template type")
	}
}

func NewEmailTemplateFromMessage(d template.Dependencies, msg Message) (EmailTemplate, error) {
	switch msg.TemplateType {
	case TypeRecoveryInvalid:
		var t email.RecoveryInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewRecoveryInvalid(d, &t), nil
	case TypeRecoveryValid:
		var t email.RecoveryValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewRecoveryValid(d, &t), nil
	case TypeRecoveryCodeInvalid:
		var t email.RecoveryCodeInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewRecoveryCodeInvalid(d, &t), nil
	case TypeRecoveryCodeValid:
		var t email.RecoveryCodeValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewRecoveryCodeValid(d, &t), nil
	case TypeVerificationInvalid:
		var t email.VerificationInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewVerificationInvalid(d, &t), nil
	case TypeVerificationValid:
		var t email.VerificationValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewVerificationValid(d, &t), nil
	case TypeVerificationCodeInvalid:
		var t email.VerificationCodeInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewVerificationCodeInvalid(d, &t), nil
	case TypeVerificationCodeValid:
		var t email.VerificationCodeValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewVerificationCodeValid(d, &t), nil
	case TypeTestStub:
		var t email.TestStubModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewTestStub(d, &t), nil
	case TypeLoginCodeValid:
		var t email.LoginCodeValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewLoginCodeValid(d, &t), nil
	case TypeRegistrationCodeValid:
		var t email.RegistrationCodeValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewRegistrationCodeValid(d, &t), nil
	default:
		return nil, errors.Errorf("received unexpected message template type: %s", msg.TemplateType)
	}
}
