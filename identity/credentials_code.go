// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"encoding/json"

	"github.com/ory/herodot"

	"github.com/pkg/errors"

	"github.com/ory/x/stringsx"
)

type CodeChannel string

const (
	CodeChannelEmail CodeChannel = AddressTypeEmail
	CodeChannelSMS   CodeChannel = AddressTypeSMS
)

func NewCodeChannel(value string) (CodeChannel, error) {
	switch f := stringsx.SwitchExact(value); {
	case f.AddCase(string(CodeChannelEmail)):
		return CodeChannelEmail, nil
	case f.AddCase(string(CodeChannelSMS)):
		return CodeChannelSMS, nil
	default:
		return "", errors.Wrap(ErrInvalidCodeAddressType, f.ToUnknownCaseErr().Error())
	}
}

// CredentialsCode represents a one time login/registration code
//
// swagger:model identityCredentialsCode
type CredentialsCode struct {
	Addresses []CredentialsCodeAddress `json:"addresses"`
}

// swagger:model identityCredentialsCodeAddress
type CredentialsCodeAddress struct {
	// The type of the address for this code
	Channel CodeChannel `json:"channel"`

	// The address for this code
	Address string `json:"address"`
}

var ErrInvalidCodeAddressType = herodot.ErrInternalServerError.WithReasonf("The address type for sending OTP codes is not supported.")

func (c *CredentialsCodeAddress) UnmarshalJSON(data []byte) (err error) {
	type alias CredentialsCodeAddress
	var ac alias
	if err := json.Unmarshal(data, &ac); err != nil {
		return err
	}

	ac.Channel, err = NewCodeChannel(string(ac.Channel))
	if err != nil {
		return err
	}

	*c = CredentialsCodeAddress(ac)
	return nil
}
