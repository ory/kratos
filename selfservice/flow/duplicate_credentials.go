// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package flow

import (
	"encoding/json"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/kratos/identity"
	"github.com/ory/x/sqlxx"
)

const internalContextDuplicateCredentialsPath = "registration_duplicate_credentials"

type DuplicateCredentialsData struct {
	CredentialsType     identity.CredentialsType
	CredentialsConfig   sqlxx.JSONRawMessage
	DuplicateIdentifier string
}

// SetDuplicateCredentials sets the duplicate credentials data in the flow's internal context.
func SetDuplicateCredentials(flow InternalContexter, creds DuplicateCredentialsData) error {
	if flow.GetInternalContext() == nil {
		flow.EnsureInternalContext()
	}
	bytes, err := sjson.SetBytes(
		flow.GetInternalContext(),
		internalContextDuplicateCredentialsPath,
		creds,
	)
	if err != nil {
		return err
	}
	flow.SetInternalContext(bytes)

	return nil
}

// DuplicateCredentials returns the duplicate credentials data from the flow's internal context.
func DuplicateCredentials(flow InternalContexter) (*DuplicateCredentialsData, error) {
	if flow.GetInternalContext() == nil {
		flow.EnsureInternalContext()
	}
	raw := gjson.GetBytes(flow.GetInternalContext(), internalContextDuplicateCredentialsPath)
	if !raw.IsObject() {
		return nil, nil
	}
	var creds DuplicateCredentialsData
	err := json.Unmarshal([]byte(raw.Raw), &creds)

	return &creds, err
}
