package node

import (
	"encoding/json"

	"github.com/ory/kratos/text"
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

type ImageAttributesModifier func(attributes *ImageAttributes)
type ImageAttributesModifiers []ImageAttributesModifier

func WithImageAttributes(f func(a *ImageAttributes)) func(a *ImageAttributes) {
	return func(a *ImageAttributes) {
		f(a)
	}
}

func applyImageAttributes(opts ImageAttributesModifiers, attributes *ImageAttributes) *ImageAttributes {
	for _, f := range opts {
		f(attributes)
	}
	return attributes
}

type ScriptAttributesModifier func(attributes *ScriptAttributes)
type ScriptAttributesModifiers []ScriptAttributesModifier

func WithScriptAttributes(f func(a *ScriptAttributes)) func(a *ScriptAttributes) {
	return func(a *ScriptAttributes) {
		f(a)
	}
}

func applyScriptAttributes(opts ScriptAttributesModifiers, attributes *ScriptAttributes) *ScriptAttributes {
	for _, f := range opts {
		f(attributes)
	}
	return attributes
}

func NewInputFieldFromJSON(name string, value interface{}, group Group, opts ...InputAttributesModifier) *Node {
	return &Node{
		Type:       Input,
		Group:      group,
		Attributes: applyInputAttributes(opts, &InputAttributes{Name: name, Type: toFormType(name, value), FieldValue: value}),
		Meta:       &Meta{},
	}
}

func NewInputField(name string, value interface{}, group Group, inputType InputAttributeType, opts ...InputAttributesModifier) *Node {
	return &Node{
		Type:       Input,
		Group:      group,
		Attributes: applyInputAttributes(opts, &InputAttributes{Name: name, Type: inputType, FieldValue: value}),
		Meta:       &Meta{},
	}
}

func NewImageField(id string, src string, group Group, opts ...ImageAttributesModifier) *Node {
	return &Node{
		Type:       Image,
		Group:      group,
		Attributes: applyImageAttributes(opts, &ImageAttributes{Source: src, Identifier: id}),
		Meta:       &Meta{},
	}
}

func NewTextField(id string, text *text.Message, group Group) *Node {
	return &Node{
		Type:       Text,
		Group:      group,
		Attributes: &TextAttributes{Text: text, Identifier: id},
		Meta:       &Meta{},
	}
}

func NewAnchorField(id string, href string, group Group, title *text.Message) *Node {
	return &Node{
		Type:       Anchor,
		Group:      group,
		Attributes: &AnchorAttributes{Title: title, HREF: href, Identifier: id},
		Meta:       &Meta{},
	}
}

func NewScriptField(name string, src string, group Group, integrity string, opts ...ScriptAttributesModifier) *Node {
	return &Node{
		Type:  Script,
		Group: group,
		Attributes: applyScriptAttributes(opts, &ScriptAttributes{
			Identifier:     name,
			Type:           "text/javascript",
			Source:         src,
			Async:          true,
			ReferrerPolicy: "no-referrer",
			CrossOrigin:    "anonymous",
			Integrity:      integrity,
			Nonce:          x.NewUUID().String(),
		}),
		Meta: &Meta{},
	}
}

func NewInputFieldFromSchema(name string, group Group, p jsonschemax.Path, opts ...InputAttributesModifier) *Node {
	attr := &InputAttributes{
		Name:     name,
		Type:     toFormType(p.Name, p.Type),
		Required: p.Required,
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

	var meta Meta
	if len(p.Title) > 0 {
		meta.Label = text.NewInfoNodeLabelGenerated(p.Title)
	}

	return &Node{
		Type:       Input,
		Attributes: applyInputAttributes(opts, attr),
		Group:      group,
		Meta:       &meta,
	}
}
