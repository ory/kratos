package form

import "github.com/ory/kratos/text"

type Form interface {
	ErrorParser
	FieldSetter
	ValueSetter
	FieldUnsetter
	MessageAdder
	CSRFSetter
	Resetter
	FieldSorter
}

// ErrorParser is capable of parsing and processing errors.
type ErrorParser interface {
	// ParseError type asserts the given error and sets the forms's errors or a
	// field's errors and if the error is not something to be handled by the
	// form itself, the error is returned for further propagation (e.g. showing a 502 status code).
	ParseError(err error) error
}

type FieldSetter interface {
	// SetField sets a field of the form.
	SetField(field Field)
}

type FieldUnsetter interface {
	// UnsetFields removes a field from the form.
	UnsetField(name string)
}

type ValueSetter interface {
	// SetValue sets a value of the form.
	SetValue(name string, value interface{})
}

type MessageAdder interface {
	// AddMessage adds a message to the form. A message can also be set for one or more fields if
	// `setForFields` is set.
	AddMessage(err *text.Message, setForFields ...string)
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
	SortFields(schemaRef string, prefix string) error
}
