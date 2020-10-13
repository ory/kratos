package x

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/x/stringsx"
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
	_, body := EasyGet(t, c, url) // nolint: bodyclose
	return body
}

func EasyCookieJar(t *testing.T, o *cookiejar.Options) *cookiejar.Jar {
	cj, err := cookiejar.New(o)
	require.NoError(t, err)
	return cj
}

func RequestURL(r *http.Request) *url.URL {
	source := *r.URL
	source.Host = stringsx.Coalesce(source.Host, r.Host)
	if source.Scheme == "" {
		source.Scheme = "https"
		if r.TLS == nil {
			source.Scheme = "http"
		}
	}

	return &source
}

func NewTransportWithHeader(h http.Header) *TransportWithHeader {
	return &TransportWithHeader{
		RoundTripper: http.DefaultTransport,
		h:            h,
	}
}

type TransportWithHeader struct {
	http.RoundTripper
	h http.Header
}

func (ct *TransportWithHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k := range ct.h {
		req.Header.Set(k, ct.h.Get(k))
	}
	return ct.RoundTripper.RoundTrip(req)
}
