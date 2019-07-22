package selfservice

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ory/x/urlx"
)

type Request struct {
	ID             string      `json:"id"`
	IssuedAt       time.Time   `json:"issued_at"`
	ExpiresAt      time.Time   `json:"expires_at"`
	RequestURL     string      `json:"request_url"`
	RequestHeaders http.Header `json:"headers"`
}

func newRequestFromHTTP(exp time.Duration, r *http.Request) *Request {
	source := urlx.Copy(r.URL)
	source.Host = r.Host

	if len(source.Scheme) == 0 {
		source.Scheme = "http"
		if r.TLS != nil {
			source.Scheme = "https"
		}
	}

	return &Request{
		ID:             uuid.New().String(),
		IssuedAt:       time.Now().UTC(),
		ExpiresAt:      time.Now().UTC().Add(exp),
		RequestURL:     source.String(),
		RequestHeaders: r.Header,
	}
}
