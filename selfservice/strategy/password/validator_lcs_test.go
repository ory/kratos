package password

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLCSLength(t *testing.T) {
	for k, tc := range []struct {
		a string
		b string
		l int
	}{
		{a: "foo", b: "foo", l: 3},
		{a: "fo", b: "foo", l: 2},
		{a: "bar", b: "foo", l: 0},
		{a: "foobar", b: "foo", l: 3},
		{a: "foobar", b: "oo", l: 2},
		{a: "foobar", b: "a", l: 1},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			require.Equal(t, tc.l, lcsLength(tc.a, tc.b))
			require.Equal(t, tc.l, lcsLength(tc.b, tc.a))
		})
	}
}
