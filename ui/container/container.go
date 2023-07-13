// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ory/kratos/ui/node"
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
	_       ErrorParser = new(Container)
	_       ValueSetter = new(Container)
	_       Resetter    = new(Container)
	_       CSRFSetter  = new(Container)
	_       NodeGetter  = new(Container)
)

// Container represents a HTML Form. The container can work with both HTTP Form and JSON requests
//
// swagger:model uiContainer
type Container struct {
	// Action should be used as the form action URL `<form action="{{ .Action }}" method="post">`.
	//
	// required: true
	Action string `json:"action" faker:"url"`

	// Method is the form method (e.g. POST)
	//
	// required: true
	Method string `json:"method" faker:"http_method"`

	// Nodes contains the form's nodes
	//
	// The form's nodes can be input fields, text, images, and other UI elements.
	//
	// required: true
	Nodes node.Nodes `json:"nodes"`

	// Messages contains all global form messages and errors.
	Messages text.Messages `json:"messages,omitempty"`
}

// New returns an empty container.
func New(action string) *Container {
	return &Container{
		Action: action,
		Method: "POST",
		Nodes:  node.Nodes{},
	}
}

// NewFromHTTPRequest creates a new Container and populates fields by parsing the HTTP Request body.
// A jsonSchemaRef needs to be added to allow HTTP Form Post Body parsing.
func NewFromHTTPRequest(r *http.Request, group node.UiNodeGroup, action string, compiler decoderx.HTTPDecoderOption) (*Container, error) {
	c := New(action)
	raw := json.RawMessage(`{}`)
	if err := decoder.Decode(r, &raw, compiler); err != nil {
		if err := c.ParseError(group, err); err != nil {
			return nil, err
		}
	}

	c.UpdateNodeValuesFromJSON(raw, "", group)
	return c, nil
}

// NewFromJSON creates a UI Container based on the provided JSON struct.
func NewFromJSON(action string, group node.UiNodeGroup, raw json.RawMessage, prefix string) *Container {
	c := New(action)
	c.UpdateNodeValuesFromJSON(raw, prefix, group)
	return c
}

// NewFromStruct creates a UI Container based on serialized contents of the provided struct.
func NewFromStruct(action string, group node.UiNodeGroup, v interface{}, prefix string) (*Container, error) {
	c := New(action)
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	c.UpdateNodeValuesFromJSON(data, prefix, group)
	return c, nil
}

// NewFromJSONSchema creates a new Container and populates the fields
// using the provided JSON Schema.
func NewFromJSONSchema(ctx context.Context, action string, group node.UiNodeGroup, jsonSchemaRef, prefix string, compiler *jsonschema.Compiler) (*Container, error) {
	c := New(action)
	nodes, err := NodesFromJSONSchema(ctx, group, jsonSchemaRef, prefix, compiler)
	if err != nil {
		return nil, err
	}

	c.Nodes = nodes
	return c, nil
}

func NodesFromJSONSchema(ctx context.Context, group node.UiNodeGroup, jsonSchemaRef, prefix string, compiler *jsonschema.Compiler) (node.Nodes, error) {
	paths, err := jsonschemax.ListPaths(ctx, jsonSchemaRef, compiler)
	if err != nil {
		return nil, err
	}

	nodes := node.Nodes{}
	for _, value := range paths {
		if value.TypeHint == jsonschemax.JSON {
			continue
		}

		name := addPrefix(value.Name, prefix, ".")
		nodes = append(nodes, node.NewInputFieldFromSchema(name, group, value))
	}

	return nodes, nil
}

func (c *Container) GetNodes() *node.Nodes {
	return &c.Nodes
}

func (c *Container) SortNodes(ctx context.Context, opts ...node.SortOption) error {
	return c.Nodes.SortBySchema(ctx, opts...)
}

// ResetMessages resets the container's own and its node's messages.
func (c *Container) ResetMessages(exclude ...string) {
	c.Messages = nil
	for k, n := range c.Nodes {
		if !stringslice.Has(exclude, n.ID()) {
			n.Messages = nil
		}
		c.Nodes[k] = n
	}
}

// Reset resets the container's errors as well as each field's value and errors.
func (c *Container) Reset(exclude ...string) {
	c.Messages = nil
	c.Nodes.Reset(exclude...)
}

// ParseError type asserts the given error and sets the container's errors or a
// field's errors and if the error is not something to be handled by the
// formUI Container, the error is returned.
//
// This method DOES NOT touch the values of the node values/names, only its errors.
func (c *Container) ParseError(group node.UiNodeGroup, err error) error {
	if e := richError(nil); errors.As(err, &e) {
		if e.StatusCode() == http.StatusBadRequest {
			c.AddMessage(group, text.NewValidationErrorGeneric(e.Reason()))
			return nil
		}
		return err
	} else if e := new(schema.ValidationError); errors.As(err, &e) {
		pointer, _ := jsonschemax.JSONPointerToDotNotation(e.InstancePtr)
		for i := range e.Messages {
			c.AddMessage(group, &e.Messages[i], pointer)
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
				c.AddMessage(group, text.NewValidationErrorRequired(segments[len(segments)-1]), pointer)
			}
		default:
			// The pointer can be ignored because if there is an error, we'll just use
			// the empty field (global error).
			causes := e.Causes
			if len(e.Causes) == 0 {
				pointer, _ := jsonschemax.JSONPointerToDotNotation(e.InstancePtr)
				c.AddMessage(group, translateValidationError(e), pointer)
				return nil
			}

			for _, ee := range causes {
				if err := c.ParseError(group, ee); err != nil {
					return err
				}
			}
		}
		return nil
	} else if e := new(schema.ValidationListError); errors.As(err, &e) {
		for _, ee := range e.Validations {
			if err := c.ParseError(group, ee); err != nil {
				return err
			}
		}
		return nil
	}
	return err
}

func translateValidationError(err *jsonschema.ValidationError) *text.Message {
	segments := strings.Split(err.SchemaPtr, "/")
	switch segments[len(segments)-1] {
	case "minLength":
		return text.NewErrorValidationMinLength(err.Message)
	case "maxLength":
		return text.NewErrorValidationMaxLength(err.Message)
	case "pattern":
		return text.NewErrorValidationInvalidFormat(err.Message)
	case "minimum":
		return text.NewErrorValidationMinimum(err.Message)
	case "exclusiveMinimum":
		return text.NewErrorValidationExclusiveMinimum(err.Message)
	case "maximum":
		return text.NewErrorValidationMaximum(err.Message)
	case "exclusiveMaximum":
		return text.NewErrorValidationExclusiveMaximum(err.Message)
	case "multipleOf":
		return text.NewErrorValidationMultipleOf(err.Message)
	case "maxItems":
		return text.NewErrorValidationMaxItems(err.Message)
	case "minItems":
		return text.NewErrorValidationMinItems(err.Message)
	case "uniqueItems":
		return text.NewErrorValidationUniqueItems(err.Message)
	case "type":
		return text.NewErrorValidationWrongType(err.Message)
	default:
		return text.NewValidationErrorGeneric(err.Message)
	}
}

// UpdateNodeValuesFromJSON sets the container's fields to the provided values.
func (c *Container) UpdateNodeValuesFromJSON(raw json.RawMessage, prefix string, group node.UiNodeGroup) {
	for k, v := range jsonx.Flatten(raw) {
		k = addPrefix(k, prefix, ".")

		if n := c.Nodes.Find(k); n != nil {
			n.Attributes.SetValue(v)
			n.Group = group
			continue
		}

		c.Nodes.Upsert(node.NewInputFieldFromJSON(k, v, group))
	}
}

// Unset removes a field from the container.
func (c *Container) UnsetNode(id string) {
	c.Nodes.Remove(id)
}

// SetCSRF sets the CSRF value using e.g. nosurf.Token(r).
func (c *Container) SetCSRF(token string) {
	c.SetNode(node.NewCSRFNode(token))
}

// SetNode sets a field.
func (c *Container) SetNode(n *node.Node) {
	c.Nodes.Upsert(n)
}

// SetValue sets a container's field to the provided name and value.
func (c *Container) SetValue(id string, n *node.Node) {
	if f := c.Nodes.Find(id); f != nil {
		f.Attributes.SetValue(n.GetValue())
		return
	}

	c.Nodes.Upsert(n)
}

// AddMessage adds the provided error, and if a non-empty names list is set,
// adds the error on the corresponding field.
func (c *Container) AddMessage(group node.UiNodeGroup, err *text.Message, setForFields ...string) {
	if len(stringslice.TrimSpaceEmptyFilter(setForFields)) == 0 {
		c.Messages = append(c.Messages, *err)
		return
	}

	for _, name := range setForFields {
		if ff := c.Nodes.Find(name); ff != nil {
			ff.Messages = append(ff.Messages, *err)
			continue
		}

		n := node.NewInputField(name, nil, node.DefaultGroup, node.InputAttributeTypeText)
		n.Messages = text.Messages{*err}
		c.Nodes = append(c.Nodes, n)
	}
}

func (c *Container) Scan(value interface{}) error {
	return sqlxx.JSONScan(c, value)
}

func (c *Container) Value() (driver.Value, error) {
	return sqlxx.JSONValue(c)
}

func addPrefix(name, prefix, separator string) string {
	if prefix == "" {
		return name
	}
	return fmt.Sprintf("%s%s%s", prefix, separator, name)
}
