package form

import (
	"net/http"
)

type CSRFGenerator func(r *http.Request) string

const CSRFTokenName = "csrf_token"

func CSRFFormFieldGenerator(g CSRFGenerator) func(r *http.Request) *Field {
	return func(r *http.Request) *Field {
		return &Field{
			Name:  CSRFTokenName,
			Type:  "hidden",
			Value: g(r),
		}
	}
}
