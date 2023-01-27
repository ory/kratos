// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package testhelpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func NewDebugClient(t *testing.T) *http.Client {
	return &http.Client{Transport: NewTransportWithLogger(http.DefaultTransport, t)}
}

func NewClientWithCookieJar(t *testing.T, jar *cookiejar.Jar, debugRedirects bool) *http.Client {
	if jar == nil {
		j, err := cookiejar.New(nil)
		jar = j
		require.NoError(t, err)
	}
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if debugRedirects {
				t.Logf("Redirect: %s", req.URL.String())
			}
			if len(via) >= 20 {
				for k, v := range via {
					t.Logf("Failed with redirect (%d): %s", k, v.URL.String())
				}
				return errors.New("stopped after 20 redirects")
			}
			return nil
		},
	}
}

func NewRequest(t *testing.T, isAPI bool, method string, url string, payload io.Reader) *http.Request {
	req, err := http.NewRequest("POST", url, payload)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")
	if isAPI {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	}

	return req
}

func NewHTTPGetJSONRequest(t *testing.T, url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	// req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("Accept", "application/json")
	return req
}

func NewHTTPGetAJAXRequest(t *testing.T, url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("Accept", "application/json")
	return req
}

func NewHTTPDeleteJSONRequest(t *testing.T, url string, in interface{}) *http.Request {
	var body bytes.Buffer
	require.NoError(t, json.NewEncoder(&body).Encode(in))
	req, err := http.NewRequest("DELETE", url, &body)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return req
}

func HTTPRequestJSON(t *testing.T, client *http.Client, method string, url string, in interface{}) ([]byte, *http.Response) {
	var body bytes.Buffer
	require.NoError(t, json.NewEncoder(&body).Encode(in))

	req, err := http.NewRequest(method, url, &body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	payload, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return payload, res
}

func HTTPPostForm(t *testing.T, client *http.Client, remote string, in *url.Values) ([]byte, *http.Response) {
	if in == nil {
		in = new(url.Values)
	}

	res, err := client.PostForm(remote, *in)
	require.NoError(t, err)
	defer res.Body.Close()

	payload, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return payload, res
}
