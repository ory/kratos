package x

import (
	"net/http"

	"github.com/ory/herodot"
)

var PseudoPanic = herodot.DefaultError{
	StatusField: http.StatusText(http.StatusInternalServerError),
	ErrorField:  "Code Bug Detected",
	ReasonField: "The code ended up at a place where it should not have. Please report this as an issue at https://github.com/ory/kratos",
	CodeField:   http.StatusConflict,
}
