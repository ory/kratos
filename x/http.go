package x

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestHTTPRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	return req
}
