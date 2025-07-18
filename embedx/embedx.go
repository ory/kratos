// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package embedx

import (
	"bytes"
	_ "embed"
	"io"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/x/configx"
	"github.com/ory/x/otelx"
)

//go:embed config.schema.json
var ConfigSchema []byte

//go:embed identity_meta.schema.json
var IdentityMetaSchema []byte

//go:embed identity_extension.schema.json
var IdentityExtensionSchema []byte

type SchemaType int

const (
	Config SchemaType = iota + 1
	IdentityMeta
	IdentityExtension
)

type Schema struct {
	id           string
	data         []byte
	dependencies []*Schema
}

var (
	identityExt = &Schema{
		id:           gjson.GetBytes(IdentityExtensionSchema, "$id").Str,
		data:         IdentityExtensionSchema,
		dependencies: nil,
	}

	schemas = map[SchemaType]*Schema{
		Config: {
			id:           gjson.GetBytes(ConfigSchema, "$id").Str,
			data:         ConfigSchema,
			dependencies: nil,
		},
		IdentityMeta: {
			id:   gjson.GetBytes(IdentityMetaSchema, "$id").Str,
			data: IdentityMetaSchema,
			dependencies: []*Schema{
				identityExt,
			},
		},
		IdentityExtension: identityExt,
	}
)

func getSchema(schema SchemaType) (*Schema, error) {
	if val, ok := schemas[schema]; ok {
		return val, nil
	}
	return nil, errors.Errorf("the specified schema type (%d) is not supported", int(schema))
}

func (s SchemaType) GetSchemaID() string {
	return schemas[s].id
}

// AddSchemaResources adds the specified schemas including their dependencies to the compiler.
// The interface is specified instead of `jsonschema.Compiler` to allow the use of any jsonschema library fork or version.
func AddSchemaResources(c interface {
	AddResource(url string, r io.Reader) error
}, opts ...SchemaType) error {
	var sc []*Schema

	for _, o := range opts {
		s, err := getSchema(o)
		if err != nil {
			return err
		}
		sc = append(sc, s)
	}

	if err := addSchemaResources(c, sc); err != nil {
		return err
	}

	if err := otelx.AddConfigSchema(c); err != nil {
		return err
	}

	return configx.AddSchemaResources(c)
}

func addSchemaResources(c interface {
	AddResource(url string, r io.Reader) error
}, schemas []*Schema) error {
	for _, s := range schemas {
		if err := c.AddResource(s.id, bytes.NewReader(s.data)); err != nil {
			return errors.WithStack(err)
		}
		if s.dependencies != nil {
			if err := addSchemaResources(c, s.dependencies); err != nil {
				return err
			}
		}
	}
	return nil
}
