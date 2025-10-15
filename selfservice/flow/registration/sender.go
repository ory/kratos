// Copyright © 2025 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package registration

import (
	"context"
	"encoding/json"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/courier/template"
	"github.com/ory/kratos/courier/template/email"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/x"
)

type senderDependencies interface {
	courier.Provider
	courier.ConfigProvider
	x.HTTPClientProvider
}

type Sender struct {
	d senderDependencies
}

func NewSender(d senderDependencies) *Sender {
	return &Sender{d: d}
}

func (s *Sender) SendDuplicateRegistrationEmail(ctx context.Context, i *identity.Identity, f *Flow) error {
	emailAddr := ""
	for _, a := range i.VerifiableAddresses {
		if a.Via == identity.AddressTypeEmail {
			emailAddr = a.Value
			break
		}
	}

	if emailAddr == "" {
		return nil
	}

	identityMap := make(map[string]interface{})
	if i.Traits != nil {
		if err := json.Unmarshal(i.Traits, &identityMap); err != nil {
			return err
		}
	}

	var transientPayload map[string]interface{}
	if f.TransientPayload != nil {
		if err := json.Unmarshal(f.TransientPayload, &transientPayload); err != nil {
			return err
		}
	}

	model := &email.RegistrationDuplicateModel{
		To:               emailAddr,
		Identity:         identityMap,
		RequestURL:       f.RequestURL,
		TransientPayload: transientPayload,
	}

	c, err := s.d.Courier(ctx)
	if err != nil {
		return err
	}

	_, err = c.QueueEmail(ctx, email.NewRegistrationDuplicate(s.d.(template.Dependencies), model))
	return err
}
