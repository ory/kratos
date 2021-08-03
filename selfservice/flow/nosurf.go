package flow

import (
	"net/http"

	"github.com/ory/kratos/x"
)

func GetCSRFToken(reg interface {
	x.CSRFProvider
	x.CSRFTokenGeneratorProvider
}, w http.ResponseWriter, r *http.Request, p Type) string {
	token := reg.GenerateCSRFToken(r)
	if p != TypeBrowser {
		return token
	}

	if token == "" {
		return reg.CSRFHandler().RegenerateToken(w, r)
	}

	return token
}
