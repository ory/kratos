// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/ory/kratos/x"
)

// CredentialsOid2 contains the configuration for credentials of the type oidc.
//
// swagger:model identityCredentialsOidc
type CredentialsOid2 struct {
	Providers []CredentialsOid2Provider `json:"providers"`
}

// CredentialsOid2 Provider contains a specific OpenID 2.0 credential for a particular connection (e.g. Steam).
//
// swagger:model identityCredentialsOid2Provider
type CredentialsOid2Provider struct {
	ClaimedId string `json:"claimed_id"`
	Provider  string `json:"provider"`
}

// NewCredentialsOid2 creates a new Open ID 2.0 credential.
func NewCredentialsOid2(claimedId, provider string) (*Credentials, error) {
	if provider == "" {
		return nil, errors.New("received empty provider in oid2 credentials")
	}

	if claimedId == "" {
		return nil, errors.New("received empty claimed ID in oid2 credentials")
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(CredentialsOid2{
		Providers: []CredentialsOid2Provider{
			{
				ClaimedId: claimedId,
				Provider:  provider,
			}},
	}); err != nil {
		return nil, errors.WithStack(x.PseudoPanic.
			WithDebugf("Unable to encode password options to JSON: %s", err))
	}

	return &Credentials{
		Type:        CredentialsTypeOID2,
		Identifiers: []string{Oid2UniqueID(provider, claimedId)},
		Config:      b.Bytes(),
	}, nil
}

func Oid2UniqueID(provider, subject string) string {
	return fmt.Sprintf("%s:%s", provider, subject)
}
