package form

import (
	"github.com/ory/kratos/ui/node"
)

type Form interface {
	NodeGetter
	ErrorParser
	//NodeSetter
	ValueSetter
	//NodeUnsetter
	CSRFSetter
	Resetter
	FieldSorter
}

// ErrorParser is capable of parsing and processing errors.
type ErrorParser interface {
	// ParseError type asserts the given error and sets the forms's errors or a
	// field's errors and if the error is not something to be handled by the
	// form itself, the error is returned for further propagation (e.g. showing a 502 status code).
	ParseError(group node.Group, err error) error
}

type NodeSetter interface {
	// SetNode sets a field of the form.
	SetNode(field node.Node)
}

type NodeUnsetter interface {
	// UnsetFields removes a field from the form.
	UnsetNode(name string)
}

type ValueSetter interface {
	// SetValue sets a value of the form.
	SetValue(name string, value *node.Node)
}

type CSRFSetter interface {
	// SetCSRF sets the CSRF value for the form.
	SetCSRF(string)
}

type Resetter interface {
	// Resets the form or field.
	Reset(exclude ...string)
}

type MessageResetter interface {
	// ResetMessages resets the form's or field's messages..
	ResetMessages(exclude ...string)
}

type FieldSorter interface {
	SortFields(schemaRef string) error
}

type NodeGetter interface {
	GetNodes() *node.Nodes
}
