package x

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/x/assertx"
	"github.com/ory/x/logrusx"

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

func TestAcceptsContentType(t *testing.T) {
	assert.True(t, AcceptsContentType(&http.Request{
		Header: map[string][]string{"Accept": {"application/json"}},
	}, "application/json"))
	assert.True(t, AcceptsContentType(&http.Request{
		Header: map[string][]string{"Accept": {"application/json; charset=utf-8"}},
	}, "application/json"))
	assert.False(t, AcceptsContentType(&http.Request{
		Header: map[string][]string{"Accept": {"application/json"}},
	}, "text/html"))
	assert.False(t, AcceptsContentType(&http.Request{
		Header: map[string][]string{"Accept": {"application/json; charset=utf-8"}},
	}, "text/html"))
	assert.True(t, AcceptsContentType(&http.Request{
		Header: map[string][]string{"Accept": {"text/html"}},
	}, "text/html"))
	assert.True(t, AcceptsContentType(&http.Request{
		Header: map[string][]string{"Accept": {"text/html, application/json;q=0.9, */*;q=0.8"}},
	}, "application/json"))
}

func TestContentNegotiationRedirection(t *testing.T) {
wr := herodot.NewJSONWriter(logrusx.New("", ""))

t.Run("case=browser", func (t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept", "text/html")

	t.Run("regular payload", func(t *testing.T) {
		w := httptest.NewRecorder()
		ContentNegotiationRedirection(w, r, json.RawMessage(`{"foo":"bar"}`), wr, "https://www.ory.sh/redir")
		loc, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "https://www.ory.sh/redir", loc.String())
	})

	t.Run("error payload", func(t *testing.T) {
		w := httptest.NewRecorder()
		ContentNegotiationRedirection(w, r, errors.New("foo"), wr, "https://www.ory.sh/redir")
		loc, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "https://www.ory.sh/redir", loc.String())
		})
	})

	t.Run("case=json", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Accept", "application/json")

		t.Run("regular payload", func(t *testing.T) {
			msg := json.RawMessage(`{"foo":"bar"}`)
			w := httptest.NewRecorder()
			ContentNegotiationRedirection(w, r, msg, wr, "https://www.ory.sh/redir")
			_, err := w.Result().Location()
			require.ErrorIs(t, err, http.ErrNoLocation)

			body := MustReadAll(w.Result().Body)
			assertx.EqualAsJSON(t, msg, json.RawMessage(body))
		})

		t.Run("error payload", func(t *testing.T) {
			ee := errors.WithStack(herodot.ErrBadRequest)
			w := httptest.NewRecorder()
			ContentNegotiationRedirection(w, r, ee, wr, "https://www.ory.sh/redir")
			_, err := w.Result().Location()
			require.ErrorIs(t, err, http.ErrNoLocation)

			body := MustReadAll(w.Result().Body)
			assertx.EqualAsJSON(t, map[string]interface{}{"error": map[string]interface{}{"code": 400, "message": "The request was malformed or contained invalid parameters", "status": "Bad Request"}}, json.RawMessage(body))
			assert.EqualValues(t, http.StatusBadRequest, w.Result().StatusCode)
		})
	})
}
