// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"fmt"

	"github.com/ory/kratos/text"
	"github.com/ory/kratos/x/webauthnx/js"
)

const (
	InputAttributeTypeText          UiNodeInputAttributeType = "text"
	InputAttributeTypePassword      UiNodeInputAttributeType = "password"
	InputAttributeTypeNumber        UiNodeInputAttributeType = "number"
	InputAttributeTypeCheckbox      UiNodeInputAttributeType = "checkbox"
	InputAttributeTypeHidden        UiNodeInputAttributeType = "hidden"
	InputAttributeTypeEmail         UiNodeInputAttributeType = "email"
	InputAttributeTypeTel           UiNodeInputAttributeType = "tel"
	InputAttributeTypeSubmit        UiNodeInputAttributeType = "submit"
	InputAttributeTypeButton        UiNodeInputAttributeType = "button"
	InputAttributeTypeDateTimeLocal UiNodeInputAttributeType = "datetime-local"
	InputAttributeTypeDate          UiNodeInputAttributeType = "date"
	InputAttributeTypeURI           UiNodeInputAttributeType = "url"
)

const (
	InputAttributeAutocompleteEmail            UiNodeInputAttributeAutocomplete = "email"
	InputAttributeAutocompleteTel              UiNodeInputAttributeAutocomplete = "tel"
	InputAttributeAutocompleteUrl              UiNodeInputAttributeAutocomplete = "url"
	InputAttributeAutocompleteCurrentPassword  UiNodeInputAttributeAutocomplete = "current-password"
	InputAttributeAutocompleteNewPassword      UiNodeInputAttributeAutocomplete = "new-password"
	InputAttributeAutocompleteOneTimeCode      UiNodeInputAttributeAutocomplete = "one-time-code"
	InputAttributeAutocompleteUsernameWebauthn UiNodeInputAttributeAutocomplete = "username webauthn"
)

// swagger:enum UiNodeInputAttributeType
type UiNodeInputAttributeType string

// swagger:enum UiNodeInputAttributeAutocomplete
type UiNodeInputAttributeAutocomplete string

// Attributes represents a list of attributes (e.g. `href="foo"` for links).
//
// swagger:model uiNodeAttributes
type Attributes interface {
	// swagger:ignore
	ID() string

	// swagger:ignore
	Reset()

	// swagger:ignore
	SetValue(value interface{})

	// swagger:ignore
	GetValue() interface{}

	// swagger:ignore
	GetNodeType() UiNodeType

	// swagger:ignore
	Matches(other Attributes) bool
}

// InputAttributes represents the attributes of an input node
//
// swagger:model uiNodeInputAttributes
type InputAttributes struct {
	// The input's element name.
	//
	// required: true
	Name string `json:"name"`

	// The input's element type.
	//
	// required: true
	Type UiNodeInputAttributeType `json:"type" faker:"-"`

	// The input's value.
	FieldValue interface{} `json:"value,omitempty" faker:"string"`

	// Mark this input field as required.
	Required bool `json:"required,omitempty"`

	// The autocomplete attribute for the input.
	Autocomplete UiNodeInputAttributeAutocomplete `json:"autocomplete,omitempty"`

	// The input's label text.
	Label *text.Message `json:"label,omitempty"`

	// The input's pattern.
	Pattern string `json:"pattern,omitempty"`

	// Sets the input's disabled field to true or false.
	//
	// required: true
	Disabled bool `json:"disabled"`

	// OnClick may contain javascript which should be executed on click. This is primarily
	// used for WebAuthn.
	//
	// Deprecated: Using OnClick requires the use of eval() which is a security risk. Use OnClickTrigger instead.
	OnClick string `json:"onclick,omitempty"`

	// OnClickTrigger may contain a WebAuthn trigger which should be executed on click.
	//
	// The trigger maps to a JavaScript function provided by Ory, which triggers actions such as PassKey registration or login.
	OnClickTrigger js.WebAuthnTriggers `json:"onclickTrigger,omitempty"`

	// OnLoad may contain javascript which should be executed on load. This is primarily
	// used for WebAuthn.
	//
	// Deprecated: Using OnLoad requires the use of eval() which is a security risk. Use OnLoadTrigger instead.
	OnLoad string `json:"onload,omitempty"`

	// OnLoadTrigger may contain a WebAuthn trigger which should be executed on load.
	//
	// The trigger maps to a JavaScript function provided by Ory, which triggers actions such as PassKey registration or login.
	OnLoadTrigger js.WebAuthnTriggers `json:"onloadTrigger,omitempty"`

	// MaxLength may contain the input's maximum length.
	MaxLength int `json:"maxlength,omitempty"`

	// NodeType represents this node's types. It is a mirror of `node.type` and
	// is primarily used to allow compatibility with OpenAPI 3.0.  In this struct it technically always is "input".
	//
	// required: true
	NodeType UiNodeType `json:"node_type"`
}

// ImageAttributes represents the attributes of an image node.
//
// swagger:model uiNodeImageAttributes
type ImageAttributes struct {
	// The image's source URL.
	//
	// format: uri
	// required: true
	Source string `json:"src"`

	// A unique identifier
	//
	// required: true
	Identifier string `json:"id"`

	// Width of the image
	//
	// required: true
	Width int `json:"width"`

	// Height of the image
	//
	// required: true
	Height int `json:"height"`

	// NodeType represents this node's types. It is a mirror of `node.type` and
	// is primarily used to allow compatibility with OpenAPI 3.0.  In this struct it technically always is "img".
	//
	// required: true
	NodeType UiNodeType `json:"node_type"`
}

// AnchorAttributes represents the attributes of an anchor node.
//
// swagger:model uiNodeAnchorAttributes
type AnchorAttributes struct {
	// The link's href (destination) URL.
	//
	// format: uri
	// required: true
	HREF string `json:"href"`

	// The link's title.
	//
	// required: true
	Title *text.Message `json:"title"`

	// A unique identifier
	//
	// required: true
	Identifier string `json:"id"`

	// NodeType represents this node's types. It is a mirror of `node.type` and
	// is primarily used to allow compatibility with OpenAPI 3.0.  In this struct it technically always is "a".
	//
	// required: true
	NodeType UiNodeType `json:"node_type"`
}

// TextAttributes represents the attributes of a text node.
//
// swagger:model uiNodeTextAttributes
type TextAttributes struct {
	// The text of the text node.
	//
	// required: true
	Text *text.Message `json:"text"`

	// A unique identifier
	//
	// required: true
	Identifier string `json:"id"`

	// NodeType represents this node's types. It is a mirror of `node.type` and
	// is primarily used to allow compatibility with OpenAPI 3.0.  In this struct it technically always is "text".
	//
	// required: true
	NodeType UiNodeType `json:"node_type"`
}

// ScriptAttributes represent script nodes which load javascript.
//
// swagger:model uiNodeScriptAttributes
type ScriptAttributes struct {
	// The script source
	//
	// required: true
	Source string `json:"src"`

	// The script async type
	//
	// required: true
	Async bool `json:"async"`

	// The script referrer policy
	//
	// required: true
	ReferrerPolicy string `json:"referrerpolicy"`

	// The script cross origin policy
	//
	// required: true
	CrossOrigin string `json:"crossorigin"`

	// The script's integrity hash
	//
	// required: true
	Integrity string `json:"integrity"`

	// The script MIME type
	//
	// required: true
	Type string `json:"type"`

	// A unique identifier
	//
	// required: true
	Identifier string `json:"id"`

	// Nonce for CSP
	//
	// A nonce you may want to use to improve your Content Security Policy.
	// You do not have to use this value but if you want to improve your CSP
	// policies you may use it. You can also choose to use your own nonce value!
	//
	// required: true
	Nonce string `json:"nonce"`

	// NodeType represents this node's types. It is a mirror of `node.type` and
	// is primarily used to allow compatibility with OpenAPI 3.0. In this struct it technically always is "script".
	//
	// required: true
	NodeType UiNodeType `json:"node_type"`
}

// DivisionAttributes represent a division section.
//
// Division sections are used for interactive widgets that require a hook in the DOM / view.
//
// swagger:model uiNodeDivisionAttributes
type DivisionAttributes struct {
	// A classname that should be rendered into the DOM.
	Classname string `json:"class,omitzero"`

	// A unique identifier
	//
	// required: true
	Identifier string `json:"id"`

	// Data is a map of key-value pairs that are passed to the division.
	//
	// They may be used for `data-...` attributes.
	Data map[string]string `json:"data,omitzero"`

	// NodeType represents this node's type. It is a mirror of `node.type` and
	// is primarily used to allow compatibility with OpenAPI 3.0. In this struct it technically always is "script".
	//
	// required: true
	NodeType UiNodeType `json:"node_type"`
}

func (d DivisionAttributes) ID() string {
	return d.Identifier
}

func (d DivisionAttributes) Reset() {}

func (d DivisionAttributes) SetValue(_ interface{}) {}

func (d DivisionAttributes) GetValue() interface{} {
	return nil
}

func (d DivisionAttributes) GetNodeType() UiNodeType {
	return d.NodeType
}

func (d DivisionAttributes) Matches(other Attributes) bool {
	ot, ok := other.(*DivisionAttributes)
	if !ok {
		return false
	}

	if len(ot.ID()) > 0 && d.ID() != ot.ID() {
		return false
	}

	return true
}

var (
	_ Attributes = new(InputAttributes)
	_ Attributes = new(ImageAttributes)
	_ Attributes = new(AnchorAttributes)
	_ Attributes = new(TextAttributes)
	_ Attributes = new(ScriptAttributes)
	_ Attributes = new(DivisionAttributes)
)

func (a *InputAttributes) ID() string {
	return a.Name
}

func (a *ImageAttributes) ID() string {
	return a.Identifier
}

func (a *AnchorAttributes) ID() string {
	return a.Identifier
}

func (a *TextAttributes) ID() string {
	return a.Identifier
}

func (a *ScriptAttributes) ID() string {
	return a.Identifier
}

func (a *InputAttributes) Matches(other Attributes) bool {
	ot, ok := other.(*InputAttributes)
	if !ok {
		return false
	}

	if len(ot.ID()) > 0 && a.ID() != ot.ID() {
		return false
	}

	if len(ot.Type) > 0 && a.Type != ot.Type {
		return false
	}

	if ot.FieldValue != nil && fmt.Sprintf("%v", a.FieldValue) != fmt.Sprintf("%v", ot.FieldValue) {
		return false
	}

	if len(ot.Name) > 0 && a.Name != ot.Name {
		return false
	}

	return true
}

func (a *ImageAttributes) Matches(other Attributes) bool {
	ot, ok := other.(*ImageAttributes)
	if !ok {
		return false
	}

	if len(ot.ID()) > 0 && a.ID() != ot.ID() {
		return false
	}

	if len(ot.Source) > 0 && a.Source != ot.Source {
		return false
	}

	return true
}

func (a *AnchorAttributes) Matches(other Attributes) bool {
	ot, ok := other.(*AnchorAttributes)
	if !ok {
		return false
	}

	if len(ot.ID()) > 0 && a.ID() != ot.ID() {
		return false
	}

	if len(ot.HREF) > 0 && a.HREF != ot.HREF {
		return false
	}

	return true
}

func (a *TextAttributes) Matches(other Attributes) bool {
	ot, ok := other.(*TextAttributes)
	if !ok {
		return false
	}

	if len(ot.ID()) > 0 && a.ID() != ot.ID() {
		return false
	}

	return true
}

func (a *ScriptAttributes) Matches(other Attributes) bool {
	ot, ok := other.(*ScriptAttributes)
	if !ok {
		return false
	}

	if len(ot.ID()) > 0 && a.ID() != ot.ID() {
		return false
	}

	if ot.Type != "" && a.Type != ot.Type {
		return false
	}

	if ot.Source != "" && a.Source != ot.Source {
		return false
	}

	return true
}

func (a *InputAttributes) SetValue(value interface{}) {
	a.FieldValue = value
}

func (a *ImageAttributes) SetValue(value interface{}) {
	a.Source, _ = value.(string)
}

func (a *AnchorAttributes) SetValue(value interface{}) {
	a.HREF, _ = value.(string)
}

func (a *TextAttributes) SetValue(value interface{}) {
	a.Text, _ = value.(*text.Message)
}

func (a *ScriptAttributes) SetValue(value interface{}) {
	a.Source, _ = value.(string)
}

func (a *InputAttributes) GetValue() interface{} {
	return a.FieldValue
}

func (a *ImageAttributes) GetValue() interface{} {
	return a.Source
}

func (a *AnchorAttributes) GetValue() interface{} {
	return a.HREF
}

func (a *TextAttributes) GetValue() interface{} {
	return a.Text
}

func (a *ScriptAttributes) GetValue() interface{} {
	return a.Source
}

func (a *InputAttributes) Reset() {
	a.FieldValue = nil
}

func (a *ImageAttributes) Reset() {
}

func (a *AnchorAttributes) Reset() {
}

func (a *TextAttributes) Reset() {
}

func (a *ScriptAttributes) Reset() {
}

func (a *InputAttributes) GetNodeType() UiNodeType {
	return UiNodeType(a.NodeType)
}

func (a *ImageAttributes) GetNodeType() UiNodeType {
	return UiNodeType(a.NodeType)
}

func (a *AnchorAttributes) GetNodeType() UiNodeType {
	return UiNodeType(a.NodeType)
}

func (a *TextAttributes) GetNodeType() UiNodeType {
	return UiNodeType(a.NodeType)
}

func (a *ScriptAttributes) GetNodeType() UiNodeType {
	return UiNodeType(a.NodeType)
}
