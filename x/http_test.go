package x

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/urlx"
)

func TestRequestURL(t *testing.T) {
	assert.EqualValues(t, RequestURL(&http.Request{
		URL: urlx.ParseOrPanic("/foo"), Host: "foobar", TLS: &tls.ConnectionState{},
	}).String(), "https://foobar/foo")
	assert.EqualValues(t, RequestURL(&http.Request{
		URL: urlx.ParseOrPanic("/foo"), Host: "foobar",
	}).String(), "http://foobar/foo")
}
