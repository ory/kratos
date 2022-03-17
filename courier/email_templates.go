package courier

import (
	"context"
	"encoding/json"

	"github.com/ory/kratos/courier/template"

	"github.com/pkg/errors"

	"github.com/ory/kratos/courier/template/email"
)

type (
	TemplateType string

	EmailTemplate interface {
		json.Marshaler
		EmailSubject(context.Context) (string, error)
		EmailBody(context.Context) (string, error)
		EmailBodyPlaintext(context.Context) (string, error)
		EmailRecipient() (string, error)
	}
)

const (
	TypeRecoveryInvalid     TemplateType = "recovery_invalid"
	TypeRecoveryValid       TemplateType = "recovery_valid"
	TypeVerificationInvalid TemplateType = "verification_invalid"
	TypeVerificationValid   TemplateType = "verification_valid"
	TypeOTP                 TemplateType = "otp"
	TypeTestStub            TemplateType = "stub"
)

func GetEmailTemplateType(t EmailTemplate) (TemplateType, error) {
	switch t.(type) {
	case *email.RecoveryInvalid:
		return TypeRecoveryInvalid, nil
	case *email.RecoveryValid:
		return TypeRecoveryValid, nil
	case *email.VerificationInvalid:
		return TypeVerificationInvalid, nil
	case *email.VerificationValid:
		return TypeVerificationValid, nil
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
	case TypeTestStub:
		var t email.TestStubModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return email.NewTestStub(d, &t), nil
	default:
		return nil, errors.Errorf("received unexpected message template type: %s", msg.TemplateType)
	}
}
