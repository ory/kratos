// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/stringsx"
	"golang.org/x/exp/slices"

	"github.com/ory/kratos/schema"
)

type SchemaExtensionCredentials struct {
	i                     *Identity
	credentialIdentifiers map[CredentialsType][]string
	l                     sync.Mutex
}

func NewSchemaExtensionCredentials(i *Identity) *SchemaExtensionCredentials {
	return &SchemaExtensionCredentials{i: i}
}

func (r *SchemaExtensionCredentials) setIdentifier(ct CredentialsType, value interface{}, addressType CredentialsIdentifierAddressType) {
	cred, ok := r.i.GetCredentials(ct)
	if !ok {
		cred = &Credentials{
			Type:        ct,
			Identifiers: []string{},
			Config:      sqlxx.JSONRawMessage{},
		}
	}
	if r.credentialIdentifiers == nil {
		r.credentialIdentifiers = make(map[CredentialsType][]string)
	}

	r.credentialIdentifiers[ct] = stringslice.Unique(append(r.credentialIdentifiers[ct], strings.ToLower(fmt.Sprintf("%s", value))))
	cred.Identifiers = r.credentialIdentifiers[ct]
	r.i.SetCredentials(ct, *cred)
}

func (r *SchemaExtensionCredentials) addIdentifierForWebAuthn(value any) {
	cred, ok := r.i.GetCredentials(CredentialsTypeWebAuthn)
	if !ok {
		cred = &Credentials{
			Type:        CredentialsTypeWebAuthn,
			Identifiers: []string{},
			Config:      sqlxx.JSONRawMessage{},
		}
	}
	if r.credentialIdentifiers == nil {
		r.credentialIdentifiers = make(map[CredentialsType][]string)
	}

	r.credentialIdentifiers[CredentialsTypeWebAuthn] = cred.Identifiers
	normalizedAddress := strings.ToLower(fmt.Sprintf("%s", value))
	if !slices.Contains(r.credentialIdentifiers[CredentialsTypeWebAuthn], normalizedAddress) {
		r.credentialIdentifiers[CredentialsTypeWebAuthn] = append(r.credentialIdentifiers[CredentialsTypeWebAuthn], normalizedAddress)
	}

	cred.Identifiers = r.credentialIdentifiers[CredentialsTypeWebAuthn]
	r.i.SetCredentials(CredentialsTypeWebAuthn, *cred)
}

func (r *SchemaExtensionCredentials) Run(ctx jsonschema.ValidationContext, s schema.ExtensionConfig, value interface{}) error {
	r.l.Lock()
	defer r.l.Unlock()

	if s.Credentials.Password.Identifier {
		r.setIdentifier(CredentialsTypePassword, value, CredentialsIdentifierAddressTypeNone)
	}

	if s.Credentials.WebAuthn.Identifier {
		r.addIdentifierForWebAuthn(value)
	}

	if s.Credentials.Code.Identifier {
		switch f := stringsx.SwitchExact(s.Credentials.Code.Via); {
		case f.AddCase(AddressTypeEmail):
			if !jsonschema.Formats["email"](value) {
				return ctx.Error("format", "%q is not a valid %q", value, s.Credentials.Code.Via)
			}

			r.setIdentifier(CredentialsTypeCodeAuth, value, AddressTypeEmail)
		// case f.AddCase(AddressTypePhone):
		// 	if !jsonschema.Formats["tel"](value) {
		// 		return ctx.Error("format", "%q is not a valid %q", value, s.Credentials.Code.Via)
		// 	}
		//
		// 	r.setIdentifier(CredentialsTypeCodeAuth, value, CredentialsIdentifierAddressType(AddressTypePhone))
		default:
			return ctx.Error("", "credentials.code.via has unknown value %q", s.Credentials.Code.Via)
		}
	}

	return nil
}

func (r *SchemaExtensionCredentials) Finish() error {
	return nil
}
