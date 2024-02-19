// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package passkey

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/schema"
	"github.com/ory/x/stringsx"
)

type SchemaExtension struct {
	WebauthnIdentifier string
	PasskeyDisplayName string
	sync.Mutex
}

func (e *SchemaExtension) Run(_ jsonschema.ValidationContext, s schema.ExtensionConfig, value any) error {
	e.Lock()
	defer e.Unlock()

	if s.Credentials.WebAuthn.Identifier {
		e.WebauthnIdentifier = strings.ToLower(fmt.Sprintf("%s", value))
	}

	if s.Credentials.Passkey.DisplayName {
		e.PasskeyDisplayName = fmt.Sprintf("%s", value)
	}

	return nil
}

func (e *SchemaExtension) Finish() error { return nil }

// PasskeyDisplayNameFromIdentity returns the passkey display name from the
// identity. It is usually the email address and used to name the passkey in the
// browser.
func (s *Strategy) PasskeyDisplayNameFromIdentity(ctx context.Context, id *identity.Identity) string {
	e := new(SchemaExtension)
	// We can ignore teh error here because proper validation happens once the identity is persisted.
	_ = s.d.IdentityValidator().ValidateWithRunner(ctx, id, e)

	return stringsx.Coalesce(e.PasskeyDisplayName, e.WebauthnIdentifier)
}

func (s *Strategy) PasskeyDisplayNameFromTraits(ctx context.Context, traits identity.Traits) string {
	id := identity.NewIdentity("")
	id.Traits = traits

	return s.PasskeyDisplayNameFromIdentity(ctx, id)
}

func (s *Strategy) PasskeyDisplayNameFromSchema(ctx context.Context, schemaURL string) (string, error) {
	ext := &passkeyDisplayNameExtension{}

	runner, err := schema.NewExtensionRunner(ctx, schema.WithCompileRunners(ext))
	if err != nil {
		return "", err
	}
	c := jsonschema.NewCompiler()
	c.ExtractAnnotations = true
	runner.Register(c)

	schem, err := c.Compile(ctx, schemaURL)
	if err != nil {
		return "", err
	}

	for key, value := range schem.Properties["traits"].Properties {
		if value.Title == ext.getLabel() {
			return "traits." + key, nil
		}
	}

	return "", errors.New("no identifier found")
}

type passkeyDisplayNameExtension struct {
	identifierLabelCandidates []string
}

func (i *passkeyDisplayNameExtension) Run(_ jsonschema.CompilerContext, config schema.ExtensionConfig, rawSchema map[string]interface{}) error {
	if config.Credentials.WebAuthn.Identifier ||
		config.Credentials.Passkey.DisplayName {
		if title, ok := rawSchema["title"]; ok {
			// The jsonschema compiler validates the title to be a string, so this should always work.
			switch t := title.(type) {
			case string:
				if t != "" {
					i.identifierLabelCandidates = append(i.identifierLabelCandidates, t)
				}
			}
		}
	}
	return nil
}

func (i *passkeyDisplayNameExtension) getLabel() string {
	if len(i.identifierLabelCandidates) != 1 {
		// sane default is set elsewhere
		return ""
	}
	return i.identifierLabelCandidates[0]
}
