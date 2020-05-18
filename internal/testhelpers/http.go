package testhelpers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func HTTPRequestJSON(t *testing.T, client *http.Client, method string, url string, in interface{}) ([]byte, *http.Response) {
	var body bytes.Buffer
	require.NoError(t, json.NewEncoder(&body).Encode(in))

	req, err := http.NewRequest(method, url, &body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	payload, err := ioutil.ReadAll(res.Body)
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

	payload, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	return payload, res
}
