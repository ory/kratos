package embedx

import (
	"bytes"
	"io"

	"github.com/tidwall/gjson"
)

import _ "embed"

//go:embed config.schema.json
var ConfigSchema string

//go:embed identity_meta.schema.json
var IdentityMetaSchema string

//go:embed identity_extension.schema.json
var IdentityExtensionSchema string

var (
	ConfigSchemaID       string
	IdentityMetaSchemaID string
)

type Schema struct {
	id       string
	data     string
	children []*Schema
}

type options struct {
	schemas []*Schema
}

func (o *options) addSchema(parentData string, children ...string) string {
	id := gjson.Get(parentData, "$id").Str
	parent := &Schema{
		id:   id,
		data: parentData,
	}

	for _, child := range children {
		parent.children = append(parent.children, &Schema{
			id:       gjson.Get(child, "$id").Str,
			data:     child,
			children: nil,
		})
	}

	o.schemas = append(o.schemas, parent)
	return id
}

type Options func(*options)

func WithConfigSchema() Options {
	return func(o *options) {
		ConfigSchemaID = o.addSchema(ConfigSchema)
	}
}

func WithIdentityMetaSchema() Options {
	return func(o *options) {
		IdentityMetaSchemaID = o.addSchema(IdentityMetaSchema, IdentityExtensionSchema)
	}
}

// AddSchemaResources adds the logging schema to the compiler.
// The interface is specified instead of `jsonschema.Compiler` to allow the use of any jsonschema library fork or version.
func AddSchemaResources(c interface {
	AddResource(url string, r io.Reader) error
}, opts ...Options) error {
	opt := &options{}

	for _, o := range opts {
		o(opt)
	}

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
		if s.children != nil {
			if err := addSchemaResources(c, s.children); err != nil {
				return err
			}
		}
	}
	return nil
}
