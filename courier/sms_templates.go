package courier

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/ory/kratos/courier/template/sms"
)

type SMSTemplate interface {
	json.Marshaler
	SMSBody(context.Context) (string, error)
	PhoneNumber() (string, error)
}

func SMSTemplateType(t SMSTemplate) (TemplateType, error) {
	switch t.(type) {
	case *sms.RecoveryMessage:
		return TypeRecoveryValidOTP, nil
	case *sms.VerificationMessage:
		return TypeVerificationValidOTP, nil
	case *sms.TestStub:
		return TypeTestStub, nil
	default:
		return "", errors.Errorf("unexpected template type")
	}
}

func NewSMSTemplateFromMessage(d Dependencies, m Message) (SMSTemplate, error) {
	switch m.TemplateType {
	case TypeRecoveryValidOTP:
		var t sms.RecoveryMessageModel
		if err := json.Unmarshal(m.TemplateData, &t); err != nil {
			return nil, err
		}
		return sms.NewRecoveryOTPMessage(d, &t), nil
	case TypeVerificationValidOTP:
		var t sms.VerificationMessageModel
		if err := json.Unmarshal(m.TemplateData, &t); err != nil {
			return nil, err
		}
		return sms.NewVerificationOTPMessage(d, &t), nil
	case TypeTestStub:
		var t sms.TestStubModel
		if err := json.Unmarshal(m.TemplateData, &t); err != nil {
			return nil, err
		}
		return sms.NewTestStub(d, &t), nil
	default:
		return nil, errors.Errorf("received unexpected message template type: %s", m.TemplateType)
	}
}
