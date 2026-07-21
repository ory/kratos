// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/ory/kratos/x"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/schema"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/stringslice"
)

type SchemaExtensionCredentials struct {
	i         *Identity
	v         map[CredentialsType][]string
	addresses []CredentialsCodeAddress
	l         sync.Mutex
}

func NewSchemaExtensionCredentials(i *Identity) *SchemaExtensionCredentials {
	return &SchemaExtensionCredentials{i: i}
}

func (r *SchemaExtensionCredentials) setIdentifier(ct CredentialsType, value interface{}) {
	cred, ok := r.i.GetCredentials(ct)
	if !ok {
		cred = &Credentials{
			Type:        ct,
			Identifiers: []string{},
			Config:      sqlxx.JSONRawMessage{},
		}
	}

	if r.v == nil {
		r.v = make(map[CredentialsType][]string)
	}

	normalized := x.GracefulNormalization(fmt.Sprintf("%s", value))

	r.v[ct] = stringslice.Unique(append(r.v[ct], normalized))
	cred.Identifiers = r.v[ct]
	r.i.SetCredentials(ct, *cred)
}

func (r *SchemaExtensionCredentials) Run(ctx jsonschema.ValidationContext, s schema.ExtensionConfig, value interface{}) error {
	r.l.Lock()
	defer r.l.Unlock()

	if s.Credentials.Password.Identifier {
		r.setIdentifier(CredentialsTypePassword, value)
	}

	if s.Credentials.WebAuthn.Identifier {
		r.setIdentifier(CredentialsTypeWebAuthn, value)
	}

	if s.Credentials.Code.Identifier {
		via, err := NewCodeChannel(s.Credentials.Code.Via)
		if err != nil {
			return ctx.Error("ory.sh~/kratos/credentials/code/via", "channel type %q must be one of %s", s.Credentials.Code.Via, strings.Join([]string{
				string(CodeChannelEmail),
				string(CodeChannelSMS),
			}, ", "))
		}

		cred := r.i.GetCredentialsOr(CredentialsTypeCodeAuth, &Credentials{
			Type:        CredentialsTypeCodeAuth,
			Identifiers: []string{},
			Config:      sqlxx.JSONRawMessage("{}"),
			Version:     1,
		})

		var conf CredentialsCode
		conf.Addresses = r.addresses

		// For the email channel with a non-strict format (e.g. `no-validate`),
		// the schema `pattern` alone gates the value and graceful normalization
		// keeps the stored address in sync with the login-code lookup. SMS
		// always needs E.164 normalization.
		var formatString string
		if raw, ok := s.RawSchema["format"]; ok {
			str, ok := raw.(string)
			if !ok {
				// Defensive: the draft-07 meta-schema already rejects a
				// non-string "format" at compile time.
				return &jsonschema.ValidationError{Message: `the "format" field must be a string`}
			}
			formatString = str
		}
		var normalized string
		if via == CodeChannelEmail && formatString != "" && formatString != "email" {
			normalized = x.GracefulNormalization(fmt.Sprintf("%s", value))
		} else {
			n, err := x.NormalizeIdentifier(fmt.Sprintf("%s", value), string(via))
			if err != nil {
				return &jsonschema.ValidationError{Message: err.Error()}
			}
			normalized = n
		}

		conf.Addresses = append(conf.Addresses, CredentialsCodeAddress{
			Channel: via,
			Address: normalized,
		})

		conf.Addresses = lo.UniqBy(conf.Addresses, func(item CredentialsCodeAddress) string {
			return fmt.Sprintf("%x:%s", item.Address, item.Channel)
		})

		sort.SliceStable(conf.Addresses, func(i, j int) bool {
			if conf.Addresses[i].Address == conf.Addresses[j].Address {
				return conf.Addresses[i].Channel < conf.Addresses[j].Channel
			}
			return conf.Addresses[i].Address < conf.Addresses[j].Address
		})

		if r.v == nil {
			r.v = make(map[CredentialsType][]string)
		}

		r.v[CredentialsTypeCodeAuth] = stringslice.Unique(append(r.v[CredentialsTypeCodeAuth],
			lo.Map(conf.Addresses, func(item CredentialsCodeAddress, _ int) string {
				return item.Address
			})...,
		))
		r.addresses = conf.Addresses

		cred.Identifiers = r.v[CredentialsTypeCodeAuth]
		cred.Config, err = json.Marshal(conf)
		if err != nil {
			return errors.WithStack(err)
		}

		r.i.SetCredentials(CredentialsTypeCodeAuth, *cred)
	}

	return nil
}

func (r *SchemaExtensionCredentials) Finish() error {
	return nil
}
