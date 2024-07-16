// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"encoding/json"

	"github.com/ory/herodot"

	"github.com/pkg/errors"

	"github.com/ory/x/stringsx"
)

type CodeAddressType string

const (
	CodeAddressTypeEmail CodeAddressType = AddressTypeEmail
	CodeAddressTypeSMS   CodeAddressType = AddressTypeSMS
)

func NewCodeAddressType(value string) (CodeAddressType, error) {
	switch f := stringsx.SwitchExact(value); {
	case f.AddCase(string(CodeAddressTypeEmail)):
		return CodeAddressTypeEmail, nil
	case f.AddCase(string(CodeAddressTypeSMS)):
		return CodeAddressTypeSMS, nil
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
	Channel CodeAddressType `json:"channel"`

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

	ac.Channel, err = NewCodeAddressType(string(ac.Channel))
	if err != nil {
		return err
	}

	*c = CredentialsCodeAddress(ac)
	return nil
}
