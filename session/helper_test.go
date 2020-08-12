package session

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBearerTokenFromRequest(t *testing.T) {
	for k, tc := range []struct {
		h http.Header
		t string
		f bool
	}{
		{
			h: http.Header{"Authorization": {"Bearer token"}},
			t: "token", f : true,
		},
		{
			h: http.Header{"Authorization": {"bearer token"}},
			t: "token", f : true,
		},
		{
			h: http.Header{"Authorization": {"beaRer token"}},
			t: "token", f : true,
		},
		{
			h: http.Header{"Authorization": {"BEARER token"}},
			t: "token", f : true,
		},
		{
			h: http.Header{"Authorization": {"notbearer token"}},
		},
		{
			h: http.Header{"Authorization": {"token"}},
		},
		{
			h: http.Header{"Authorization": {}},
		},
		{
			h: http.Header{},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			token, found := bearerTokenFromRequest(&http.Request{Header: tc.h})
			assert.Equal(t, tc.f, found)
			assert.Equal(t, tc.t, token)
		})
	}
}
