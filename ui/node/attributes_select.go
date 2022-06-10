package node

import (
	"fmt"

	"github.com/ory/kratos/text"
	"github.com/ory/x/jsonschemax"
)

type SelectAttributesModifier func(attributes *SelectAttributes)
type SelectAttributesModifiers []SelectAttributesModifier

func WithRequiredSelectAttribute(a *SelectAttributes) {
	a.Required = true
}

func WithSelectAttributes(f func(a *SelectAttributes)) func(a *SelectAttributes) {
	return func(a *SelectAttributes) {
		f(a)
	}
}

func applySelectAttributes(opts []SelectAttributesModifier, attributes *SelectAttributes) *SelectAttributes {
	for _, f := range opts {
		f(attributes)
	}
	return attributes
}

func NewSelectFieldFromJSON(name string, value interface{}, group UiNodeGroup, options []SelectAttributeOption, opts ...SelectAttributesModifier) *Node {
	return &Node{
		Type:       Select,
		Group:      group,
		Attributes: applySelectAttributes(opts, &SelectAttributes{Name: name, Options: options, FieldValue: value}),
		Meta:       &Meta{},
	}
}

func NewSelectField(name string, value interface{}, group UiNodeGroup, options []SelectAttributeOption, opts ...SelectAttributesModifier) *Node {
	return &Node{
		Type:       Select,
		Group:      group,
		Attributes: applySelectAttributes(opts, &SelectAttributes{Name: name, Options: options, FieldValue: value}),
		Meta:       &Meta{},
	}
}

func NewSelectFieldFromSchema(name string, group UiNodeGroup, p jsonschemax.Path, opts ...SelectAttributesModifier) *Node {
	attr := &SelectAttributes{
		Name:     name,
		Required: p.Required,
	}

	// Set disabled if the custom property is set
	if isDisabled, ok := p.CustomProperties[DisableFormField]; ok {
		if isDisabled, ok := isDisabled.(bool); ok {
			attr.Disabled = isDisabled
		}
	}

	var meta Meta

	var options []SelectAttributeOption

	for _, value := range p.Enum {
		valueAsString := fmt.Sprint(value)

		options = append(options, SelectAttributeOption{
			Label: valueAsString,
			Value: valueAsString,
		})
	}

	attr.Options = options

	if len(p.Title) > 0 {
		meta.Label = text.NewInfoNodeLabelGenerated(p.Title)
	}

	return &Node{
		Type:       Select,
		Attributes: applySelectAttributes(opts, attr),
		Group:      group,
		Meta:       &meta,
	}
}
