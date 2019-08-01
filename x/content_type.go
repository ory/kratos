package x

import (
	"mime"
	"net/http"
	"strings"
)

// Determine whether the request `content-type` includes a
// server-acceptable mime-type
//
// Failure should yield an HTTP 415 (`http.StatusUnsupportedMediaType`)
func HasContentType(r *http.Request, mimetype string) bool {
	contentType := r.Header.Get("Content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(strings.TrimSpace(v))
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}
