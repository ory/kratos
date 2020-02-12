package x

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestHTTPRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)
	return req
}

func EasyGet(t *testing.T, c *http.Client, url string) (*http.Response, []byte) {
	res, err := c.Get(url)
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	return res, body
}

func EasyGetBody(t *testing.T, c *http.Client, url string) []byte {
	_, body := EasyGet(t, c, url)
	return body
}

func EasyCookieJar(t *testing.T, o *cookiejar.Options) *cookiejar.Jar {
	cj, err := cookiejar.New(o)
	require.NoError(t, err)
	return cj
}
