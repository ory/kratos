package courier

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/ory/kratos/courier/template"
)

type (
	TemplateType  string
	EmailTemplate interface {
		json.Marshaler
		EmailSubject() (string, error)
		EmailBody() (string, error)
		EmailBodyPlaintext() (string, error)
		EmailRecipient() (string, error)
	}
)

const (
	TypeRecoveryInvalid     TemplateType = "recovery_invalid"
	TypeRecoveryValid       TemplateType = "recovery_valid"
	TypeVerificationInvalid TemplateType = "verification_invalid"
	TypeVerificationValid   TemplateType = "verification_valid"
	TypeTestStub            TemplateType = "stub"
)

func GetTemplateType(t EmailTemplate) (TemplateType, error) {
	switch t.(type) {
	case *template.RecoveryInvalid:
		return TypeRecoveryInvalid, nil
	case *template.RecoveryValid:
		return TypeRecoveryValid, nil
	case *template.VerificationInvalid:
		return TypeVerificationInvalid, nil
	case *template.VerificationValid:
		return TypeVerificationValid, nil
	case *template.TestStub:
		return TypeTestStub, nil
	default:
		return "", errors.Errorf("unexpected template type")
	}
}

func NewEmailTemplateFromMessage(c SMTPConfig, msg Message) (EmailTemplate, error) {
	switch msg.TemplateType {
	case TypeRecoveryInvalid:
		var t template.RecoveryInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return template.NewRecoveryInvalid(c, &t), nil
	case TypeRecoveryValid:
		var t template.RecoveryValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return template.NewRecoveryValid(c, &t), nil
	case TypeVerificationInvalid:
		var t template.VerificationInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return template.NewVerificationInvalid(c, &t), nil
	case TypeVerificationValid:
		var t template.VerificationValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return template.NewVerificationValid(c, &t), nil
	case TypeTestStub:
		var t template.TestStubModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return template.NewTestStub(c, &t), nil
	default:
		return nil, errors.Errorf("received unexpected message template type: %s", msg.TemplateType)
	}
}
