package settings_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func TestRequest(t *testing.T) {
	alice := x.NewUUID()
	malice := x.NewUUID()
	for k, tc := range []struct {
		r         *settings.Request
		s         *session.Session
		expectErr bool
	}{
		{
			r: settings.NewRequest(
				time.Hour,
				&http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"},
				&session.Session{Identity: &identity.Identity{ID: alice}},
			),
			s: &session.Session{Identity: &identity.Identity{ID: alice}},
		},
		{
			r: settings.NewRequest(
				time.Hour,
				&http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"},
				&session.Session{Identity: &identity.Identity{ID: alice}},
			),
			s:         &session.Session{Identity: &identity.Identity{ID: malice}},
			expectErr: true,
		},
		{
			r: settings.NewRequest(
				-time.Hour,
				&http.Request{URL: urlx.ParseOrPanic("http://foo/bar/baz"), Host: "foo"},
				&session.Session{Identity: &identity.Identity{ID: alice}},
			),
			s:         &session.Session{Identity: &identity.Identity{ID: alice}},
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
