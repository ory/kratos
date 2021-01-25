package node

import (
	"encoding/json"

	"github.com/ory/kratos/x"
	"github.com/ory/x/jsonschemax"
)

const DisableFormField = "disableFormField"

func toFormType(n string, i interface{}) InputAttributeType {
	switch n {
	case x.CSRFTokenName:
		return InputAttributeTypeHidden
	case "password":
		return InputAttributeTypePassword
	}

	switch i.(type) {
	case float64, int64, int32, float32, json.Number:
		return InputAttributeTypeNumber
	case bool:
		return InputAttributeTypeCheckbox
	}

	return InputAttributeTypeText
}

type InputAttributesModifier func(attributes *InputAttributes)
type InputAttributesModifiers []InputAttributesModifier

func WithRequiredInputAttribute(a *InputAttributes) {
	a.Required = true
}

func WithInputAttributes(f func(a *InputAttributes)) func(a *InputAttributes) {
	return func(a *InputAttributes) {
		f(a)
	}
}

func applyInputAttributes(opts []InputAttributesModifier, attributes *InputAttributes) *InputAttributes {
	for _, f := range opts {
		f(attributes)
	}
	return attributes
}

func NewInputFieldFromJSON(name string, value interface{}, group Group, opts ...InputAttributesModifier) *Node {
	return &Node{
		Type: Input, Group: group,
		Attributes: applyInputAttributes(opts, &InputAttributes{Name: name, Type: toFormType(name, value), FieldValue: value}),
	}
}

func NewInputField(name string, value interface{}, group Group, inputType InputAttributeType, opts ...InputAttributesModifier) *Node {
	return &Node{
		Type: Input, Group: group,
		Attributes: applyInputAttributes(opts, &InputAttributes{Name: name, Type: inputType, FieldValue: value}),
	}
}

func NewInputFieldFromSchema(name string, group Group, p jsonschemax.Path, opts ...InputAttributesModifier) *Node {
	attr := &InputAttributes{
		Name: name,
		Type: toFormType(p.Name, p.Type),
	}

	// If format is set, we can make a more distinct decision:
	switch p.Format {
	case "date-time":
		attr.Type = InputAttributeTypeDateTimeLocal
	case "email":
		attr.Type = InputAttributeTypeEmail
	case "date":
		attr.Type = InputAttributeTypeDate
	case "uri":
		attr.Type = InputAttributeTypeURI
	case "regex":
		attr.Type = InputAttributeTypeText
	}

	// Other properties
	if p.Pattern != nil {
		attr.Pattern = p.Pattern.String()
	}

	// Set disabled if the custom property is set
	if isDisabled, ok := p.CustomProperties[DisableFormField]; ok {
		if isDisabled, ok := isDisabled.(bool); ok {
			attr.Disabled = isDisabled
		}
	}

	return &Node{Type: Input, Attributes: applyInputAttributes(opts, attr), Group: group}
}
