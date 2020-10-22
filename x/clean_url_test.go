package x

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"
)

func TestCleanPath(t *testing.T) {
	n := negroni.New(CleanPath)
	n.UseHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.String()))
	})
	ts := httptest.NewServer(n)
	defer ts.Close()

	for k, tc := range [][]string{
		{"//foo", "/foo"},
		{"//foo//bar", "/foo/bar"},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			res, err := ts.Client().Get(ts.URL + tc[0])
			require.NoError(t, err)
			defer res.Body.Close()
			body, err := ioutil.ReadAll(res.Body)
			assert.Equal(t, string(body), tc[1])
		})
	}
}
