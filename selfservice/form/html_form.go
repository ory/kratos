package form

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/ory/x/sqlxx"

	"github.com/ory/jsonschema/v3"

	"github.com/ory/x/decoderx"
	"github.com/ory/x/jsonschemax"
	"github.com/ory/x/jsonx"
	"github.com/ory/x/stringslice"

	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/text"
)

var (
	decoder             = decoderx.NewHTTP()
	_       ErrorParser = new(HTMLForm)
	_       ValueSetter = new(HTMLForm)
	_       Resetter    = new(HTMLForm)
	_       CSRFSetter  = new(HTMLForm)
)

// HTMLForm represents a HTML Form. The container can work with both HTTP Form and JSON requests
//
// swagger:model form
type HTMLForm struct {
	sync.RWMutex `faker:"-"`

	// Action should be used as the form action URL `<form action="{{ .Action }}" method="post">`.
	//
	// required: true
	Action string `json:"action" faker:"url"`

	// Method is the form method (e.g. POST)
	//
	// required: true
	Method string `json:"method" faker:"http_method"`

	// Fields contains the form fields.
	//
	// required: true
	Fields Fields `json:"fields"`

	// Messages contains all global form messages and errors.
	Messages text.Messages `json:"messages,omitempty"`
}

// NewHTMLForm returns an empty container.
func NewHTMLForm(action string) *HTMLForm {
	return &HTMLForm{
		Action: action,
		Method: "POST",
		Fields: Fields{},
	}
}

// NewHTMLFormFromRequestBody creates a new HTMLForm and populates fields by parsing the HTTP Request body.
// A jsonSchemaRef needs to be added to allow HTTP Form Post Body parsing.
func NewHTMLFormFromRequestBody(r *http.Request, action string, compiler decoderx.HTTPDecoderOption) (*HTMLForm, error) {
	c := NewHTMLForm(action)
	raw := json.RawMessage(`{}`)
	if err := decoder.Decode(r, &raw, compiler); err != nil {
		if err := c.ParseError(err); err != nil {
			return nil, err
		}
	}

	for k, v := range jsonx.Flatten(raw) {
		c.SetValue(k, v)
	}

	return c, nil
}

// NewHTMLFormFromJSON creates a HTML form based on the provided JSON struct.
func NewHTMLFormFromJSON(action string, raw json.RawMessage, prefix string) *HTMLForm {
	c := NewHTMLForm(action)
	c.SetValuesFromJSON(raw, prefix)

	return c
}

// NewHTMLFormFromJSONSchema creates a new HTMLForm and populates the fields
// using the provided JSON Schema.
func NewHTMLFormFromJSONSchema(action, jsonSchemaRef, prefix string, compiler *jsonschema.Compiler) (*HTMLForm, error) {
	paths, err := jsonschemax.ListPaths(jsonSchemaRef, compiler)
	if err != nil {
		return nil, err
	}

	c := NewHTMLForm(action)
	for _, value := range paths {
		name := addPrefix(value.Name, prefix, ".")
		c.Fields = append(c.Fields, fieldFromPath(name, value))
	}

	return c, nil
}

func (c *HTMLForm) SortFields(schemaRef string) error {
	sortFunc, err := c.Fields.sortBySchema(schemaRef, "")
	if err != nil {
		return err
	}

	sort.SliceStable(c.Fields, sortFunc)
	return nil
}

// Reset resets the container's errors as well as each field's value and errors.
func (c *HTMLForm) ResetMessages(exclude ...string) {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	c.Messages = nil
	for k, f := range c.Fields {
		if !stringslice.Has(exclude, f.Name) {
			f.Messages = nil
		}
		c.Fields[k] = f
	}
}

// Reset resets the container's errors as well as each field's value and errors.
func (c *HTMLForm) Reset(exclude ...string) {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	c.Messages = nil
	for k, f := range c.Fields {
		if !stringslice.Has(exclude, f.Name) {
			f.Reset()
		}
		c.Fields[k] = f
	}
}

// ParseError type asserts the given error and sets the container's errors or a
// field's errors and if the error is not something to be handled by the
// form container, the error is returned.
//
// This method DOES NOT touch the values of the form fields, only its errors.
func (c *HTMLForm) ParseError(err error) error {
	c.defaults()
	if e := richError(nil); errors.As(err, &e) {
		if e.StatusCode() == http.StatusBadRequest {
			c.AddMessage(text.NewValidationErrorGeneric(e.Reason()))
			return nil
		}
		return err
	} else if e := new(schema.ValidationError); errors.As(err, &e) {
		pointer, _ := jsonschemax.JSONPointerToDotNotation(e.InstancePtr)
		for i := range e.Messages {
			c.AddMessage(&e.Messages[i], pointer)
		}
		return nil
	} else if e := new(jsonschema.ValidationError); errors.As(err, &e) {
		switch ctx := e.Context.(type) {
		case *jsonschema.ValidationErrorContextRequired:
			for _, required := range ctx.Missing {
				// The pointer can be ignored because if there is an error, we'll just use
				// the empty field (global error).
				pointer, _ := jsonschemax.JSONPointerToDotNotation(required)
				segments := strings.Split(required, "/")
				c.AddMessage(text.NewValidationErrorRequired(segments[len(segments)-1]), pointer)
			}
		default:
			// The pointer can be ignored because if there is an error, we'll just use
			// the empty field (global error).
			for _, ee := range e.Causes {
				pointer, _ := jsonschemax.JSONPointerToDotNotation(ee.InstancePtr)
				c.AddMessage(text.NewValidationErrorGeneric(ee.Message), pointer)
			}
		}
		return nil
	}
	return err
}

// SetValues sets the container's fields to the provided values.
func (c *HTMLForm) SetValues(values map[string]interface{}) {
	c.defaults()
	for k, v := range values {
		c.SetValue(k, v)
	}
}

// SetValuesFromJSON sets the container's fields to the provided values.
func (c *HTMLForm) SetValuesFromJSON(raw json.RawMessage, prefix string) {
	c.defaults()
	for k, v := range jsonx.Flatten(raw) {
		if prefix != "" {
			k = prefix + "." + k
		}
		c.SetValue(k, v)
	}
}

// getField returns a pointer to the field with the given name.
func (c *HTMLForm) getField(name string) *Field {
	// to prevent blocks we don't use c.defaults() here
	if c.Fields == nil {
		return nil
	}

	for i := range c.Fields {
		if c.Fields[i].Name == name {
			return &c.Fields[i]
		}
	}

	return nil
}

// SetRequired sets the container's fields required.
func (c *HTMLForm) SetRequired(fields ...string) {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	for _, field := range fields {
		if f := c.getField(field); f != nil {
			f.Required = true
		}
	}
}

// Unset removes a field from the container.
func (c *HTMLForm) UnsetField(name string) {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	for i := range c.Fields {
		if c.Fields[i].Name == name {
			c.Fields = append(c.Fields[:i], c.Fields[i+1:]...)
			return
		}
	}
}

// SetCSRF sets the CSRF value using e.g. nosurf.Token(r).
func (c *HTMLForm) SetCSRF(token string) {
	c.SetField(Field{
		Name:     CSRFTokenName,
		Type:     "hidden",
		Required: true,
		Value:    token,
	})
}

// SetField sets a field.
func (c *HTMLForm) SetField(field Field) {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	for i := range c.Fields {
		if c.Fields[i].Name == field.Name {
			c.Fields[i] = field
			return
		}
	}

	c.Fields = append(c.Fields, field)
}

// SetValue sets a container's field to the provided name and value.
func (c *HTMLForm) SetValue(name string, value interface{}) {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	if f := c.getField(name); f != nil {
		f.Value = value
		return
	}
	c.Fields = append(c.Fields, Field{
		Name:  name,
		Value: value,
		Type:  toFormType(name, value),
	})
}

// AddMessage adds the provided error, and if a non-empty names list is set,
// adds the error on the corresponding field.
func (c *HTMLForm) AddMessage(err *text.Message, setForFields ...string) {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	if len(stringslice.TrimSpaceEmptyFilter(setForFields)) == 0 {
		c.Messages = append(c.Messages, *err)
		return
	}

	for _, name := range setForFields {
		if ff := c.getField(name); ff != nil {
			ff.Messages = append(ff.Messages, *err)
			continue
		}

		c.Fields = append(c.Fields, Field{
			Name:     name,
			Messages: text.Messages{*err},
		})
	}
}

func (c *HTMLForm) Scan(value interface{}) error {
	return sqlxx.JSONScan(c, value)
}
func (c *HTMLForm) Value() (driver.Value, error) {
	return sqlxx.JSONValue(c)
}

func (c *HTMLForm) defaults() {
	c.Lock()
	defer c.Unlock()
	if c.Fields == nil {
		c.Fields = Fields{}
	}
}
