package form

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

type ValueSetter interface {
	// SetValue sets a value of the form.
	SetValue(name string, value interface{})
}

type ErrorAdder interface {
	// AddError adds an error to the form.
	AddError(err *Error, names ...string)
}

type CSRFSetter interface {
	// SetCSRF sets the CSRF value for the form.
	SetCSRF(string)
}

type Resetter interface {
	// Reset resets errors.
	Reset()
}

type FieldSorter interface {
	SortFields(schemaRef string, prefix string) error
}
