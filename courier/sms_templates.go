// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
	case *sms.OTPMessage:
		return TypeOTP, nil
	case *sms.TestStub:
		return TypeTestStub, nil
	default:
		return "", errors.Errorf("unexpected template type")
	}
}

func NewSMSTemplateFromMessage(d Dependencies, m Message) (SMSTemplate, error) {
	switch m.TemplateType {
	case TypeOTP:
		var t sms.OTPMessageModel
		if err := json.Unmarshal(m.TemplateData, &t); err != nil {
			return nil, err
		}
		return sms.NewOTPMessage(d, &t), nil
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
