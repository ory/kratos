package form

type (
	richError interface {
		StatusCode() int
		Reason() string
	}

	// swagger:model formErrors
	Errors []Error

	// swagger:model formError
	Error struct {
		// Code    FormErrorCode `json:"id,omitempty"`
		Message string `json:"message"`
		// FieldName string `json:"field_name,omitempty"`
	}
)
