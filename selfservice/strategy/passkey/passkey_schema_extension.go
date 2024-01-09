// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/schema"
)

type SchemaExtension struct {
	Identifier string
	sync.Mutex
}

func (e *SchemaExtension) Run(_ jsonschema.ValidationContext, s schema.ExtensionConfig, value any) error {
	e.Lock()
	defer e.Unlock()

	if s.Credentials.Passkey.Identifier {
		e.Identifier = strings.ToLower(fmt.Sprintf("%s", value))
	}

	return nil
}

func (e *SchemaExtension) Finish() error { return nil }
