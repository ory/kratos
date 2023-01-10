// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package embedx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/jsonschema/v3"
)

func TestAddSchemaResources(t *testing.T) {

	for _, tc := range []struct {
		description       string
		dependencies      []SchemaType
		extraDependencies []SchemaType
		mustFail          bool
		failMessage       string
	}{
		{
			description:       "config schema",
			dependencies:      []SchemaType{Config},
			extraDependencies: nil,
		},
		{
			description:       "identity schema with dependencies",
			dependencies:      []SchemaType{IdentityMeta},
			extraDependencies: []SchemaType{IdentityExtension},
		},
		{
			description:       "multiple schemas",
			dependencies:      []SchemaType{Config, IdentityMeta, IdentityExtension},
			extraDependencies: nil,
		},
		{
			description:       "verify dependencies are also loaded",
			dependencies:      []SchemaType{IdentityMeta},
			extraDependencies: []SchemaType{IdentityExtension},
		},
		{
			description:  "must fail on unsupported schema types",
			dependencies: []SchemaType{4, 10},
			mustFail:     true,
			failMessage:  "the specified schema type (4) is not supported",
		},
	} {
		t.Run("case="+tc.description, func(t *testing.T) {
			c := jsonschema.NewCompiler()
			err := AddSchemaResources(c, tc.dependencies...)

			if tc.mustFail {
				assert.Errorf(t, err, "an error must be thrown on `%s`", tc.description)
				assert.EqualError(t, err, tc.failMessage)
			} else {
				assert.NoError(t, err)
			}

			if !tc.mustFail {
				for _, s := range append(tc.dependencies, tc.extraDependencies...) {
					_, err := c.Compile(context.Background(), s.GetSchemaID())
					assert.NoError(t, err)
				}
			}

		})
	}
}
