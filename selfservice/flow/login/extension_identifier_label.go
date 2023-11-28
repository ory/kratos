// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"
	"sort"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/schema"
)

type identifierLabelExtension struct {
	identifierLabelCandidates []string
}

var _ schema.CompileExtension = new(identifierLabelExtension)

func GetIdentifierLabelFromSchema(ctx context.Context, schemaURL string) (string, error) {
	ext := &identifierLabelExtension{}

	runner, err := schema.NewExtensionRunner(ctx, schema.WithCompileRunners(ext))
	if err != nil {
		return "", err
	}
	c := jsonschema.NewCompiler()
	runner.Register(c)

	_, err = c.Compile(ctx, schemaURL)
	if err != nil {
		return "", err
	}
	return ext.getLabel(), nil
}

func (i *identifierLabelExtension) Run(_ jsonschema.CompilerContext, config schema.ExtensionConfig, rawSchema map[string]interface{}) error {
	if config.Credentials.Password.Identifier ||
		config.Credentials.WebAuthn.Identifier ||
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
	if len(i.identifierLabelCandidates) == 0 {
		// sane default is set elsewhere
		return ""
	}
	// sort the candidates to get a deterministic result
	sort.Strings(i.identifierLabelCandidates)
	// just take the first, no good way to decide which one is the best
	return i.identifierLabelCandidates[0]
}
