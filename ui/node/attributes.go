package node

import "github.com/ory/kratos/text"

const (
	InputAttributeTypeText          InputAttributeType = "text"
	InputAttributeTypePassword      InputAttributeType = "password"
	InputAttributeTypeNumber        InputAttributeType = "number"
	InputAttributeTypeCheckbox      InputAttributeType = "checkbox"
	InputAttributeTypeHidden        InputAttributeType = "hidden"
	InputAttributeTypeEmail         InputAttributeType = "email"
	InputAttributeTypeSubmit         InputAttributeType = "submit"
	InputAttributeTypeDateTimeLocal InputAttributeType = "datetime-local"
	InputAttributeTypeDate          InputAttributeType = "date"
	InputAttributeTypeURI           InputAttributeType = "url"
)

type (
	InputAttributeType string

	// Attributes represents a list of attributes (e.g. `href="foo"` for links).
	Attributes interface {
		ID() string
		Reset()
		SetValue(value interface{})
		GetValue() interface{}
	}

	// InputAttributes represents the attributes of an input node
	InputAttributes struct {
		// The input's element name.
		Name string `json:"name"`

		// The input's element type.
		Type InputAttributeType `json:"type" faker:"-"`

		// The input's value.
		FieldValue interface{} `json:"value" faker:"string"`

		// Mark this input field as required.
		Required bool `json:"required"`

		// The input's label text.
		Label *text.Message `json:"label,omitempty"`

		// The input's pattern.
		Pattern string `json:"pattern,omitempty"`

		// Sets the input's disabled field to true or false.
		Disabled bool `json:"disabled"`
	}

	// ImageAttributes represents the attributes of an image node
	ImageAttributes struct {
		// The image's source URL.
		//
		// format: uri
		// required: true
		Source string `json:"src"`
	}

	// AnchorAttributes represents the attributes of an anchor node
	AnchorAttributes struct {
		// The link's href (destination) URL.
		//
		// format: uri
		// required: true
		HREF string `json:"href"`

		// The link's title.
		//
		// required: true
		Title *text.Message `json:"title"`
	}

	// TextAttributes represents the attributes of a text node
	TextAttributes struct {
		// The text of the text node.
		//
		// required: true
		Text *text.Message `json:"text"`
	}
)

var (
	_ Attributes = new(InputAttributes)
	_ Attributes = new(ImageAttributes)
	_ Attributes = new(AnchorAttributes)
	_ Attributes = new(TextAttributes)
)

func (a *InputAttributes) ID() string {
	return a.Name
}

func (a *ImageAttributes) ID() string {
	return ""
}

func (a *AnchorAttributes) ID() string {
	return ""
}

func (a *TextAttributes) ID() string {
	return ""
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

func (a *InputAttributes) Reset() {
	a.FieldValue = nil
}

func (a *ImageAttributes) Reset() {
}

func (a *AnchorAttributes) Reset() {
}

func (a *TextAttributes) Reset() {
}
