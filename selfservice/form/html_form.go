package form

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/ory/x/errorsx"
	"github.com/santhosh-tekuri/jsonschema/v2"

	"github.com/ory/x/decoderx"
	"github.com/ory/x/jsonschemax"
	"github.com/ory/x/jsonx"
	"github.com/ory/x/stringslice"

	"github.com/ory/kratos/schema"
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
	sync.RWMutex

	// Action should be used as the form action URL (<form action="{{ .Action }}" method="post">).
	Action string `json:"action"`

	// Method is the form method (e.g. POST)
	Method string `json:"method"`

	// Fields contains the form fields asdfasdffasd
	Fields Fields `json:"fields"`

	// Errors contains all form errors. These will be duplicates of the individual field errors.
	Errors []Error `json:"errors,omitempty"`
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
	if err := decoder.Decode(r, &raw, compiler,
		decoderx.HTTPDecoderSetIgnoreParseErrorsStrategy(decoderx.ParseErrorIgnore),
	); err != nil {
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
	for k, v := range jsonx.Flatten(raw) {
		if prefix != "" {
			k = prefix + "." + k
		}
		c.SetValue(k, v)
	}
	return c
}

// NewHTMLFormFromJSONSchema creates a new HTMLForm and populates the fields
// using the provided JSON Schema.
func NewHTMLFormFromJSONSchema(action, jsonSchemaRef, prefix string) (*HTMLForm, error) {
	paths, err := jsonschemax.ListPaths(jsonSchemaRef, nil)
	if err != nil {
		return nil, err
	}

	c := NewHTMLForm(action)
	for _, value := range paths {
		name := addPrefix(value.Name, prefix, ".")
		c.Fields[name] = Field{
			Name: name,
			Type: toFormType(value.Name, value.Type),
		}
	}

	return c, nil
}

// Reset resets the container's errors as well as each field's value and errors.
func (c *HTMLForm) Reset() {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	c.Errors = nil
	for k, f := range c.Fields {
		f.Reset()
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
	switch e := errorsx.Cause(err).(type) {
	case richError:
		if e.StatusCode() == http.StatusBadRequest {
			c.AddError(&Error{Message: e.Reason()})
			return nil
		}
		return err
	case *jsonschema.ValidationError:
		for _, err := range append([]*jsonschema.ValidationError{e}, e.Causes...) {
			pointer, _ := jsonschemax.JSONPointerToDotNotation(err.InstancePtr)
			if err.Context == nil {
				// The pointer can be ignored because if there is an error, we'll just use
				// the empty field (global error).
				c.AddError(&Error{Message: err.Message}, pointer)
				continue
			}
			switch ctx := err.Context.(type) {
			case *jsonschema.ValidationContextRequired:
				for _, required := range ctx.Missing {
					// The pointer can be ignored because if there is an error, we'll just use
					// the empty field (global error).
					pointer, _ := jsonschemax.JSONPointerToDotNotation(required)
					c.AddError(&Error{Message: err.Message}, pointer)
				}
			default:
				c.AddError(&Error{Message: err.Message}, pointer)
				continue
			}
		}
		return nil
	case schema.ResultErrors:
		for _, ei := range e {
			switch ei.Type() {
			case "invalid_credentials":
				c.AddError(&Error{Message: ei.Description()})
			default:
				c.AddError(&Error{Message: ei.String()}, ei.Field())
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

// SetRequired sets the container's fields required.
func (c *HTMLForm) SetRequired(fields ...string) {
	c.defaults()
	for _, field := range fields {
		ff, ok := c.Fields[field]
		if !ok {
			continue
		}
		ff.Required = true
	}
}

// Unset removes a field from the container.
func (c *HTMLForm) Unset(name string) {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	delete(c.Fields, name)
}

// SetCSRF sets the CSRF value using e.g. nosurf.Token(r).
func (c *HTMLForm) SetCSRF(token string) {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	c.Fields[CSRFTokenName] = Field{
		Name:     CSRFTokenName,
		Type:     "hidden",
		Required: true,
		Value:    token,
	}
}

// SetField sets a field.
func (c *HTMLForm) SetField(name string, field Field) {
	c.defaults()
	c.Lock()
	defer c.Unlock()
	c.Fields[name] = field
}

// SetValue sets a container's field to the provided name and value.
func (c *HTMLForm) SetValue(name string, value interface{}) {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	ff, ok := c.Fields[name]
	if !ok {
		c.Fields[name] = Field{
			Name:  name,
			Type:  toFormType(name, value),
			Value: value,
		}
	}

	ff.Name = name
	ff.Value = value
	if len(ff.Type) == 0 {
		ff.Type = toFormType(name, value)
	}
	c.Fields[name] = ff
}

// AddError adds the provided error, and if a non-empty names list is set,
// adds the error on the corresponding field.
func (c *HTMLForm) AddError(err *Error, names ...string) {
	c.defaults()
	c.Lock()
	defer c.Unlock()

	if len(stringslice.TrimSpaceEmptyFilter(names)) == 0 {
		c.Errors = append(c.Errors, *err)
		return
	}

	for _, name := range names {
		ff, ok := c.Fields[name]
		if !ok {
			c.Fields[name] = Field{
				Name:   name,
				Errors: []Error{*err},
			}
			continue
		}

		ff.Errors = append(ff.Errors, *err)
		c.Fields[name] = ff
	}
}

func (c *HTMLForm) defaults() {
	c.Lock()
	defer c.Unlock()
	if c.Fields == nil {
		c.Fields = Fields{}
	}
}
