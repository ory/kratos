// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/knadh/koanf/parsers/json"
)

type router interface {
	GET(path string, handle httprouter.Handle)
}

func NewConfigHashHandler(c Provider, router router) {
	router.GET("/health/config", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/plain")
		if revision := c.Config().GetProvider(r.Context()).String("revision"); len(revision) > 0 {
			_, _ = fmt.Fprintf(w, "%s", revision)
		} else {
			bytes, _ := c.Config().GetProvider(r.Context()).Marshal(json.Parser())
			_, _ = fmt.Fprintf(w, "%x", sha256.Sum256(bytes))
		}
	})
}
