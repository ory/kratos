package form

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ory/x/jsonschemax"

	"github.com/ory/kratos/schema"
)

// Fields contains multiple fields
//
// swagger:model formFields
type Fields []Field

// Field represents a HTML Form Field
//
// swagger:model formField
type Field struct {
	// Name is the equivalent of <input name="{{.Name}}">
	//
	// required: true
	Name string `json:"name"`
	// Type is the equivalent of <input type="{{.Type}}">
	//
	// required: true
	Type string `json:"type"`

	// Pattern is the equivalent of <input pattern="{{.Pattern}}">
	Pattern string `json:"pattern,omitempty"`

	// Disabled is the equivalent of <input disabled="{{.Disabled}}">
	Disabled bool `json:"disabled,omitempty"`

	// Required is the equivalent of <input required="{{.Required}}">
	//
	// required: true
	Required bool `json:"required,omitempty"`

	// Value is the equivalent of <input value="{{.Value}}">
	Value interface{} `json:"value,omitempty" faker:"name"`

	// Errors contains all validation errors this particular field has caused.
	Errors []Error `json:"errors,omitempty"`
}

// Reset resets a field's value and errors.
func (f *Field) Reset() {
	f.Errors = nil
	f.Value = nil
}

func (ff *Fields) sortBySchema(schemaRef, prefix string) (func(i, j int) bool, error) {
	schemaKeys, err := schema.GetKeysInOrder(schemaRef)
	if err != nil {
		return nil, err
	}
	keysInOrder := []string{
		CSRFTokenName,
		"identifier",
		"password",
	}
	for _, k := range schemaKeys {
		keysInOrder = append(keysInOrder, fmt.Sprintf("%s.%s", prefix, k))
	}

	getKeyPosition := func(name string) int {
		lastPrefix := len(keysInOrder)
		for i, n := range keysInOrder {
			if strings.HasPrefix(name, n) {
				lastPrefix = i
			}
		}
		return lastPrefix
	}

	return func(i, j int) bool {
		return getKeyPosition((*ff)[i].Name) < getKeyPosition((*ff)[j].Name)
	}, nil
}

func toFormType(n string, i interface{}) string {
	switch n {
	case CSRFTokenName:
		return "hidden"
	case "password":
		return "password"
	}

	switch i.(type) {
	case float64, int64, int32, float32, json.Number:
		return "number"
	case bool:
		return "checkbox"
	}

	return "text"
}

func fieldFromPath(name string, p jsonschemax.Path) Field {
	f := Field{
		Name: name,
		Type: "text",
	}

	// Estimating type
	f.Type = toFormType(p.Name, p.Type)

	switch p.Format {
	case "date-time":
		f.Type = "datetime-local"
	case "email":
		f.Type = "email"
	case "date":
		f.Type = "date"
	case "uri":
		f.Type = "url"
	case "regex":
		f.Type = "text"
	}

	// Other properties
	if p.Pattern != nil {
		f.Pattern = p.Pattern.String()
	}

	return f
}

func addPrefix(name, prefix, separator string) string {
	if prefix == "" {
		return name
	}
	return fmt.Sprintf("%s%s%s", prefix, separator, name)
}
