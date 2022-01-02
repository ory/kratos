package node

import "github.com/ory/kratos/text"

const (
	InputAttributeTypeText          InputAttributeType = "text"
	InputAttributeTypePassword      InputAttributeType = "password"
	InputAttributeTypeNumber        InputAttributeType = "number"
	InputAttributeTypeCheckbox      InputAttributeType = "checkbox"
	InputAttributeTypeHidden        InputAttributeType = "hidden"
	InputAttributeTypeEmail         InputAttributeType = "email"
	InputAttributeTypeSubmit        InputAttributeType = "submit"
	InputAttributeTypeButton        InputAttributeType = "button"
	InputAttributeTypeDateTimeLocal InputAttributeType = "datetime-local"
	InputAttributeTypeDate          InputAttributeType = "date"
	InputAttributeTypeURI           InputAttributeType = "url"
)

// swagger:model uiNodeInputAttributeType
type InputAttributeType string

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
	GetNodeType() Type
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
	Type InputAttributeType `json:"type" faker:"-"`

	// The input's value.
	FieldValue interface{} `json:"value,omitempty" faker:"string"`

	// Mark this input field as required.
	Required bool `json:"required,omitempty"`

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
	OnClick string `json:"onclick,omitempty"`

	// NodeType represents this node's types. It is a mirror of `node.type` and
	// is primarily used to allow compatibility with OpenAPI 3.0.
	//
	// required: true
	NodeType Type `json:"node_type"`
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
	Width int `json:"width,omitempty"`

	// Height of the image
	Height int `json:"height,omitempty"`

	// NodeType represents this node's types. It is a mirror of `node.type` and
	// is primarily used to allow compatibility with OpenAPI 3.0.
	//
	// required: true
	NodeType Type `json:"node_type"`
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
	// is primarily used to allow compatibility with OpenAPI 3.0.
	//
	// required: true
	NodeType Type `json:"node_type"`
}

// TextAttributes represents the attributes of a text node.
//
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
	// is primarily used to allow compatibility with OpenAPI 3.0.
	//
	// required: true
	NodeType Type `json:"node_type"`
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
	// is primarily used to allow compatibility with OpenAPI 3.0.
	//
	// required: true
	NodeType Type `json:"node_type"`
}

var (
	_ Attributes = new(InputAttributes)
	_ Attributes = new(ImageAttributes)
	_ Attributes = new(AnchorAttributes)
	_ Attributes = new(TextAttributes)
	_ Attributes = new(ScriptAttributes)
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

func (a *InputAttributes) GetNodeType() Type {
	return a.NodeType
}

func (a *ImageAttributes) GetNodeType() Type {
	return a.NodeType
}

func (a *AnchorAttributes) GetNodeType() Type {
	return a.NodeType
}

func (a *TextAttributes) GetNodeType() Type {
	return a.NodeType
}

func (a *ScriptAttributes) GetNodeType() Type {
	return a.NodeType
}
