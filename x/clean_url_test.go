// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, string(body), tc[1])
		})
	}
}
