package continuity

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerOptions(t *testing.T) {
	for k, tc := range []struct {
		err  bool
		e    func(t *testing.T, actual *managerOptions)
		opts []ManagerOption
	}{
		{
			e: func(t *testing.T, actual *managerOptions) {
				assert.EqualValues(t, time.Minute, actual.ttl)
			},
		},
		{
			opts: []ManagerOption{WithLifespan(time.Minute * 5)},
			e: func(t *testing.T, actual *managerOptions) {
				assert.EqualValues(t, time.Minute*5, actual.ttl)
			},
		},
		{
			opts: []ManagerOption{WithPayload(map[string]interface{}{"foo": "bar"})},
			e: func(t *testing.T, actual *managerOptions) {
				assert.EqualValues(t, map[string]interface{}{"foo": "bar"}, actual.payloadRaw)
				assert.JSONEq(t, `{"foo":"bar"}`, string(actual.payload))
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			o, err := newManagerOptions(tc.opts)
			if tc.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tc.e(t, o)
		})
	}
}
