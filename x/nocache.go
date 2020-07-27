package x

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// NoCache adds `Cache-Control: 0` to the response header.
func NoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "0")
}

// NoCacheHandler wraps httprouter.Handle with `Cache-Control: 0` headers.
func NoCacheHandler(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		NoCache(w)
		handle(w, r, ps)
	}
}
