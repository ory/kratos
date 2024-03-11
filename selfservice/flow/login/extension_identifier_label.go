// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/ory/herodot"
	"github.com/ory/kratos/text"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/schema"
)

type identifierLabelExtension struct {
	field                     string
	identifierLabelCandidates []string
}

var (
	_               schema.CompileExtension = new(identifierLabelExtension)
	ErrUnknownTrait                         = herodot.ErrInternalServerError.WithReasonf("Trait does not exist in identity schema")
)

func GetIdentifierLabelFromSchema(ctx context.Context, schemaURL string) (*text.Message, error) {
	return GetIdentifierLabelFromSchemaWithField(ctx, schemaURL, "")
}

func GetIdentifierLabelFromSchemaWithField(ctx context.Context, schemaURL string, trait string) (*text.Message, error) {
	ext := &identifierLabelExtension{
		field: trait,
	}

	runner, err := schema.NewExtensionRunner(ctx, schema.WithCompileRunners(ext))
	if err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	c.ExtractAnnotations = true
	runner.Register(c)

	s, err := c.Compile(ctx, schemaURL)
	if err != nil {
		return nil, err
	}

	if trait != "" {
		f, ok := s.Properties["traits"].Properties[trait]
		if !ok {
			knownTraits := lo.Keys(s.Properties["traits"].Properties)
			return nil, errors.WithStack(ErrUnknownTrait.WithDetail("trait", trait).WithDetail("known_traits", knownTraits))
		}
		return text.NewInfoNodeLabelGenerated(f.Title), nil
	}

	metaLabel := text.NewInfoNodeLabelID()
	if label := ext.getLabel(); label != "" {
		metaLabel = text.NewInfoNodeLabelGenerated(label)
	}
	return metaLabel, nil
}

func (i *identifierLabelExtension) Run(_ jsonschema.CompilerContext, config schema.ExtensionConfig, rawSchema map[string]interface{}) error {
	if config.Credentials.Password.Identifier ||
		config.Credentials.WebAuthn.Identifier ||
		config.Credentials.Passkey.DisplayName ||
		config.Credentials.TOTP.AccountName ||
		config.Credentials.Code.Identifier {
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

func (i *identifierLabelExtension) getLabel() string {
	if len(i.identifierLabelCandidates) != 1 {
		// sane default is set elsewhere
		return ""
	}
	return i.identifierLabelCandidates[0]
}
