package container

import (
	"github.com/ory/kratos/ui/node"
)

// ErrorParser is capable of parsing and processing errors.
type ErrorParser interface {
	// ParseError type asserts the given error and sets the forms's errors or a
	// field's errors and if the error is not something to be handled by the
	// formUI Container itself, the error is returned for further propagation (e.g. showing a 502 status code).
	ParseError(group node.Group, err error) error
}

type NodeSetter interface {
	// SetNode sets (adds / replaces) a node.
	SetNode(field node.Node)
}

type NodeUnsetter interface {
	// UnsetFields removes a node.
	UnsetNode(name string)
}

type ValueSetter interface {
	// SetValue sets a value the passed node.
	SetValue(name string, value *node.Node)
}

type CSRFSetter interface {
	// SetCSRF sets the CSRF value.
	SetCSRF(string)
}

type Resetter interface {
	// Resets all values and messages recursively.
	Reset(exclude ...string)
}

type MessageResetter interface {
	// ResetMessages resets the messages recursively.
	ResetMessages(exclude ...string)
}

type FieldSorter interface {
	SortNodes(schemaRef string, prefix string, keysInOrder []string) error
}

type NodeGetter interface {
	GetNodes() *node.Nodes
}
