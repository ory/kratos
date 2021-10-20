package embedx

import (
	"bytes"
	"io"

	"github.com/pkg/errors"

	"github.com/tidwall/gjson"

	_ "embed"
)

//go:embed config.schema.json
var ConfigSchema string

//go:embed identity_meta.schema.json
var IdentityMetaSchema string

//go:embed identity_extension.schema.json
var IdentityExtensionSchema string

type SchemaType int

const (
	Config SchemaType = iota
	IdentityMeta
	IdentityExtension
)

type Schema struct {
	id           string
	data         string
	dependencies []*Schema
}

var (
	identityExt = &Schema{
		id:           gjson.Get(IdentityExtensionSchema, "$id").Str,
		data:         IdentityExtensionSchema,
		dependencies: nil,
	}

	schemas = map[SchemaType]*Schema{
		Config: {
			id:           gjson.Get(ConfigSchema, "$id").Str,
			data:         ConfigSchema,
			dependencies: nil,
		},
		IdentityMeta: {
			id:   gjson.Get(IdentityMetaSchema, "$id").Str,
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

	return nil
}

func addSchemaResources(c interface {
	AddResource(url string, r io.Reader) error
}, schemas []*Schema) error {
	for _, s := range schemas {
		if err := c.AddResource(s.id, bytes.NewBufferString(s.data)); err != nil {
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
