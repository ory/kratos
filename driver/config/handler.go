package config

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/knadh/koanf/parsers/json"
)

func NewConfigHashHandler(c Provider, router *httprouter.Router) {
	router.GET("/health/config", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		bytes, _ := c.Config(r.Context()).Source().Marshal(json.Parser())
		sum := sha256.Sum256(bytes)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintf(w, "%x", sum)
	})
}
