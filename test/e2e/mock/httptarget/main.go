// Copyright © 2024 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"cmp"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/julienschmidt/httprouter"

	"github.com/ory/graceful"
)

var (
	documentsLock sync.RWMutex
	documents     = make(map[string][]byte)
)

func main() {
	port := cmp.Or(os.Getenv("PORT"), "4471")
	server := graceful.WithDefaults(&http.Server{Addr: fmt.Sprintf(":%s", port)})
	register(server)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		log.Fatalln(err)
	}
}

func register(server *http.Server) {
	router := httprouter.New()

	router.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = w.Write([]byte("OK"))
	})

	router.GET("/documents/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")

		documentsLock.RLock()
		doc, ok := documents[id]
		documentsLock.RUnlock()

		if ok {
			_, _ = w.Write(doc)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	router.PUT("/documents/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		documentsLock.Lock()
		defer documentsLock.Unlock()
		id := ps.ByName("id")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		documents[id] = body
	})

	router.DELETE("/documents/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		documentsLock.Lock()
		defer documentsLock.Unlock()
		id := ps.ByName("id")

		delete(documents, id)
		w.WriteHeader(http.StatusNoContent)
	})

	server.Handler = router
}
