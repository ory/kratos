package recovery_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/session"
)

func TestRequest(t *testing.T) {
	u := &http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"}
	for k, tc := range []struct {
		r         *recovery.Request
		s         *session.Session
		expectErr bool
	}{
		{
			r: recovery.NewRequest(time.Hour, "", u),
		},
		{
			r:         recovery.NewRequest(-time.Hour, "", u),
			expectErr: true,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			err := tc.r.Valid(tc.s)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
