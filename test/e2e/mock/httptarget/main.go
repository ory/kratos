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
)

var (
	documentsLock sync.RWMutex
	documents     = make(map[string][]byte)
)

func main() {
	port := cmp.Or(os.Getenv("PORT"), "4471")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func init() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	http.HandleFunc("GET /documents/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		documentsLock.RLock()
		doc, ok := documents[id]
		documentsLock.RUnlock()

		if ok {
			_, _ = w.Write(doc)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	http.HandleFunc("PUT /documents/{id}", func(w http.ResponseWriter, r *http.Request) {
		documentsLock.Lock()
		defer documentsLock.Unlock()
		id := r.PathValue("id")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		documents[id] = body
	})

	http.HandleFunc("DELETE /documents/{id}", func(w http.ResponseWriter, r *http.Request) {
		documentsLock.Lock()
		defer documentsLock.Unlock()
		id := r.PathValue("id")

		delete(documents, id)
		w.WriteHeader(http.StatusNoContent)
	})
}
