package form

import (
	"encoding/json"
	"fmt"
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
	Name string `json:"name"`
	// Type is the equivalent of <input type="{{.Type}}">
	Type string `json:"type,omitempty"`
	// Required is the equivalent of <input required="{{.Required}}">
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

func (ff Fields) Len() int {
	return len(ff)
}

func (ff Fields) Swap(i, j int) {
	ff[i], ff[j] = ff[j], ff[i]
}

func (ff Fields) Less(i, j int) bool {
	return ff[i].Name < ff[j].Name
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

func addPrefix(name, prefix, separator string) string {
	if prefix == "" {
		return name
	}
	return fmt.Sprintf("%s%s%s", prefix, separator, name)
}
