package x

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/ory/x/urlx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetermineReturnToURL(t *testing.T) {
	for k, tc := range []struct {
		isTLS     bool
		expect    func(t *testing.T, got string, ts *httptest.Server)
		expectErr bool
		path      string
	}{
		{
			path:      "?return_to=/foo",
			expectErr: false,
			expect: func(t *testing.T, got string, ts *httptest.Server) {
				assert.Equal(t, ts.URL+"/foo", got)
			},
		},
		{
			path:      "",
			expectErr: false,
			expect: func(t *testing.T, got string, ts *httptest.Server) {
				assert.Equal(t, ts.URL, got)
			},
		},
		{
			path:      "?return_to=https://ory.sh/asdf",
			expectErr: false,
			expect: func(t *testing.T, got string, ts *httptest.Server) {
				assert.Equal(t, "https://ory.sh/asdf", got)
			},
		},
		{
			path:      "?return_to=http://ory.sh/asdf",
			expectErr: true,
		},
		{
			path:      "?return_to=https://not-ory.sh/asdf",
			expectErr: true,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(1)

			var ts *httptest.Server
			f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer wg.Done()
				ru, err := DetermineReturnToURL(r.URL, urlx.ParseOrPanic(ts.URL), []url.URL{*urlx.ParseOrPanic("https://ory.sh/")})
				if tc.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					tc.expect(t, ru, ts)
				}

				w.WriteHeader(http.StatusNoContent)
			})

			if tc.isTLS {
				ts = httptest.NewServer(f)
			} else {
				ts = httptest.NewTLSServer(f)
			}

			defer ts.Close()

			res, err := ts.Client().Get(ts.URL + "/" + tc.path)
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusNoContent, res.StatusCode)

			wg.Wait()
		})
	}
}
