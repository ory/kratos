// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package totp

import (
	"fmt"
	"sync"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/schema"
)

type SchemaExtension struct {
	AccountName string
	l           sync.Mutex
}

func NewSchemaExtension(fallback string) *SchemaExtension {
	return &SchemaExtension{AccountName: fallback}
}

func (r *SchemaExtension) Run(_ jsonschema.ValidationContext, s schema.ExtensionConfig, value interface{}) error {
	r.l.Lock()
	defer r.l.Unlock()
	if s.Credentials.TOTP.AccountName {
		r.AccountName = fmt.Sprintf("%s", value)
	}
	return nil
}

func (r *SchemaExtension) Finish() error {
	return nil
}
