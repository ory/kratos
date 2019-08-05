package selfservice

import (
	"net/http"
)

type CSRFGenerator func(r *http.Request) string

const CSRFTokenName = "csrf_token"

func CSRFFormFieldGenerator(g CSRFGenerator) func(r *http.Request) *FormField {
	return func(r *http.Request) *FormField {
		return &FormField{
			Name:  CSRFTokenName,
			Type:  "hidden",
			Value: g(r),
		}
	}
}
