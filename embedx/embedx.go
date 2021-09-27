package embedx

import (
	"bytes"
	"io"

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
type SchemaID string

const (
	Config SchemaType = iota
	IdentityMeta
	IdentityExtension
)

var (
	ConfigSchemaID            SchemaID
	IdentityMetaSchemaID      SchemaID
	IdentityExtensionSchemaID SchemaID
)

type Schema struct {
	id           string
	data         string
	dependencies []*Schema
}

type options struct {
	schemas []*Schema
}

func (o *options) assemble(schemas ...SchemaType) {
	for _, schema := range schemas {
		parent, dependencies := schema.GetEmbeds()
		sc, parentId := addSchema(parent, dependencies...)
		schema.initSchemaID(parentId)
		o.schemas = append(o.schemas, sc)
	}
}

func addSchema(parentData string, dependencies ...string) (*Schema, SchemaID) {
	id := gjson.Get(parentData, "$id").Str
	parent := &Schema{
		id:   id,
		data: parentData,
	}

	for _, d := range dependencies {
		parent.dependencies = append(parent.dependencies, &Schema{
			id:           gjson.Get(d, "$id").Str,
			data:         d,
			dependencies: nil,
		})
	}

	return parent, SchemaID(id)
}

func (s SchemaType) getSchemaDetails() (*SchemaID, string, []string) {
	switch s {
	case Config:
		return &ConfigSchemaID, ConfigSchema, nil
	case IdentityMeta:
		return &IdentityMetaSchemaID, IdentityMetaSchema, []string{IdentityExtensionSchema}
	case IdentityExtension:
		return &IdentityExtensionSchemaID, IdentityExtensionSchema, nil
	}
	return nil, "", nil
}

func (s SchemaType) initSchemaID(schemaId SchemaID) {
	switch s {
	case Config:
		ConfigSchemaID = schemaId
	case IdentityMeta:
		IdentityMetaSchemaID = schemaId
	case IdentityExtension:
		IdentityExtensionSchemaID = schemaId
	}
}

func (s SchemaType) GetSchemaID() *SchemaID {
	id, _, _ := s.getSchemaDetails()
	return id
}

func (s SchemaType) GetEmbeds() (string, []string) {
	_, schemaEmbed, schemaDependencies := s.getSchemaDetails()
	return schemaEmbed, schemaDependencies
}

func (s *SchemaID) ToString() string {
	return string(*s)
}

// AddSchemaResources adds the logging schema to the compiler.
// The interface is specified instead of `jsonschema.Compiler` to allow the use of any jsonschema library fork or version.
func AddSchemaResources(c interface {
	AddResource(url string, r io.Reader) error
}, opts ...SchemaType) error {

	opt := &options{}
	opt.assemble(opts...)

	if err := addSchemaResources(c, opt.schemas); err != nil {
		return err
	}

	return nil
}

func addSchemaResources(c interface {
	AddResource(url string, r io.Reader) error
}, schemas []*Schema) error {
	for _, s := range schemas {
		if err := c.AddResource(s.id, bytes.NewBufferString(s.data)); err != nil {
			return err
		}
		if s.dependencies != nil {
			if err := addSchemaResources(c, s.dependencies); err != nil {
				return err
			}
		}
	}
	return nil
}
