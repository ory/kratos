package container

import (
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
func NewFromHTTPRequest(r *http.Request, group node.Group, action string, compiler decoderx.HTTPDecoderOption) (*Container, error) {
	c := New(action)
	raw := json.RawMessage(`{}`)
	if err := decoder.Decode(r, &raw, compiler); err != nil {
		if err := c.ParseError(group, err); err != nil {
			return nil, err
		}
	}

	c.UpdateNodesFromJSON(raw, "", group)
	return c, nil
}

// NewFromJSON creates a UI Container based on the provided JSON struct.
func NewFromJSON(action string, group node.Group, raw json.RawMessage, prefix string) *Container {
	c := New(action)
	c.UpdateNodesFromJSON(raw, prefix, group)
	return c
}

// NewFromJSONSchema creates a new Container and populates the fields
// using the provided JSON Schema.
func NewFromJSONSchema(action string, group node.Group, jsonSchemaRef, prefix string, compiler *jsonschema.Compiler) (*Container, error) {
	c := New(action)
	nodes, err := NodesFromJSONSchema(group, jsonSchemaRef, prefix, compiler)
	if err != nil {
		return nil, err
	}

	c.Nodes = nodes
	return c, nil
}

func NodesFromJSONSchema(group node.Group, jsonSchemaRef, prefix string, compiler *jsonschema.Compiler) (node.Nodes, error) {
	paths, err := jsonschemax.ListPaths(jsonSchemaRef, compiler)
	if err != nil {
		return nil, err
	}

	nodes := node.Nodes{}
	for _, value := range paths {
		name := addPrefix(value.Name, prefix, ".")
		nodes = append(nodes, node.NewInputFieldFromSchema(name, group, value))
	}

	return nodes, nil
}

func (c *Container) GetNodes() *node.Nodes {
	return &c.Nodes
}

func (c *Container) SortNodes(opts ...node.SortOption) error {
	return c.Nodes.SortBySchema(opts...)
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
func (c *Container) ParseError(group node.Group, err error) error {
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
			var causes = e.Causes
			if len(e.Causes) == 0 {
				pointer, _ := jsonschemax.JSONPointerToDotNotation(e.InstancePtr)
				c.AddMessage(group, text.NewValidationErrorGeneric(e.Message), pointer)
				return nil
			}

			for _, ee := range causes {
				if err := c.ParseError(group, ee); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return err
}

// UpdateNodesFromJSON sets the container's fields to the provided values.
func (c *Container) UpdateNodesFromJSON(raw json.RawMessage, prefix string, group node.Group) {
	for k, v := range jsonx.Flatten(raw) {
		if prefix != "" {
			k = prefix + "." + k
		}

		if n := c.Nodes.Find(k); n != nil {
			n.Attributes.SetValue(v)
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
func (c *Container) AddMessage(group node.Group, err *text.Message, setForFields ...string) {
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
