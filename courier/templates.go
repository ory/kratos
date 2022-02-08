package courier

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/ory/kratos/courier/template"
)

type (
	TemplateType  string
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

func NewEmailTemplateFromMessage(d SMTPDependencies, msg Message) (EmailTemplate, error) {
	switch msg.TemplateType {
	case TypeRecoveryInvalid:
		var t template.RecoveryInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return template.NewRecoveryInvalid(d, &t), nil
	case TypeRecoveryValid:
		var t template.RecoveryValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return template.NewRecoveryValid(d, &t), nil
	case TypeVerificationInvalid:
		var t template.VerificationInvalidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return template.NewVerificationInvalid(d, &t), nil
	case TypeVerificationValid:
		var t template.VerificationValidModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return template.NewVerificationValid(d, &t), nil
	case TypeTestStub:
		var t template.TestStubModel
		if err := json.Unmarshal(msg.TemplateData, &t); err != nil {
			return nil, err
		}
		return template.NewTestStub(d, &t), nil
	default:
		return nil, errors.Errorf("received unexpected message template type: %s", msg.TemplateType)
	}
}
