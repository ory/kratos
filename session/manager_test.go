package session_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/herodot"

	"github.com/ory/kratos/session"
)

func TestErrAALNotSatisfied_PassReturnToParameter(t *testing.T) {
	cases := []struct {
		name       string
		instance   *session.ErrAALNotSatisfied
		requestURL string
		wantErr    assert.ErrorAssertionFunc
		expected   string
	}{
		{
			name: "no parameter",
			instance: &session.ErrAALNotSatisfied{
				DefaultError: &herodot.DefaultError{},
				RedirectTo:   "https://localhost/?foo=bar",
			},
			requestURL: "https://localhost:1234/",
			wantErr:    assert.NoError,
			expected:   "https://localhost/?foo=bar",
		},
		{
			name: "pass return_to parameter",
			instance: &session.ErrAALNotSatisfied{
				DefaultError: &herodot.DefaultError{},
				RedirectTo:   "https://localhost/?foo=bar",
			},
			requestURL: "https://localhost:1234/?return_to=https%3A%2F%2Fory.sh",
			wantErr:    assert.NoError,
			expected:   "https://localhost/?foo=bar&return_to=https%3A%2F%2Fory.sh",
		},
		{
			name: "invalid RedirectTo URL",
			instance: &session.ErrAALNotSatisfied{
				DefaultError: &herodot.DefaultError{},
				RedirectTo:   "https://user:{{{@localhost/?foo=bar",
			},
			requestURL: "https://localhost:1234/?return_to=https%3A%2F%2Fory.sh",
			wantErr:    assert.Error,
		},
		{
			name: "invalid request URL URL",
			instance: &session.ErrAALNotSatisfied{
				DefaultError: &herodot.DefaultError{},
				RedirectTo:   "https://localhost/?foo=bar",
			},
			requestURL: "https://user:{{{@localhost:1234/?return_to=https%3A%2F%2Fory.sh",
			wantErr:    assert.Error,
		},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("case=%s", tc.name), func(t *testing.T) {
			err := tc.instance.PassReturnToParameter(tc.requestURL)

			tc.wantErr(t, err)
			if err == nil {
				assert.Equal(t, tc.expected, tc.instance.RedirectTo)
			}
		})
	}
}
