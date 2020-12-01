package x

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// NoCache adds `Cache-Control: private, no-cache, no-store, must-revalidate` to the response header.
func NoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
}

// NoCacheHandler wraps httprouter.Handle with `Cache-Control: private, no-cache, no-store, must-revalidate` headers.
func NoCacheHandler(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		NoCache(w)
		handle(w, r, ps)
	}
}
