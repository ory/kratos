// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package login

import (
	"context"

	"github.com/ory/kratos/text"
	"github.com/ory/x/jsonschemax"

	"github.com/ory/jsonschema/v3"
	"github.com/ory/kratos/schema"
)

func GetIdentifierLabelFromSchema(ctx context.Context, schemaURL string) (*text.Message, error) {
	runner, err := schema.NewExtensionRunner(ctx)
	if err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	c.ExtractAnnotations = true
	runner.Register(c)

	paths, err := jsonschemax.ListPaths(ctx, schemaURL, c)
	if err != nil {
		return nil, err
	}

	labels := []jsonschemax.Path{}
	for _, path := range paths {
		if ext := path.CustomProperties[schema.ExtensionName]; ext != nil {
			config, ok := ext.(*schema.ExtensionConfig)
			if !ok {
				continue
			}
			if config.Credentials.Password.Identifier ||
				config.Credentials.WebAuthn.Identifier ||
				config.Credentials.Passkey.DisplayName ||
				config.Credentials.TOTP.AccountName ||
				config.Credentials.Code.Identifier {
				labels = append(labels, path)
			}
		}
	}

	metaLabel := text.NewInfoNodeLabelID()
	if len(labels) == 1 && labels[0].Title != "" {
		metaLabel = text.NewInfoNodeLabelGenerated(labels[0].Title, labels[0].Name)
	}
	return metaLabel, nil
}
