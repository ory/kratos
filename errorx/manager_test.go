package errorx

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
)

func TestManager(t *testing.T) {
	l := logrus.New()
	var managers = map[string]Manager{
		"memory": NewMemoryManager(l, herodot.NewJSONWriter(l), nil),
	}

	for n, m := range managers {
		t.Run(fmt.Sprintf("manager=%s", n), func(t *testing.T) {
			sent := herodot.ErrNotFound.WithReason("foobar")
			id, err := m.Add(sent)
			require.NoError(t, err)

			got, err := m.Read(id)
			require.NoError(t, err)
			assert.EqualValues(t, []error{sent}, got)

			require.NoError(t, m.Clear(time.Second, false))
			got, err = m.Read(id)
			require.NoError(t, err)

			gotu, _ := json.Marshal(got)
			sentu, _ := json.Marshal([]error{sent})

			assert.JSONEq(t, string(sentu), string(gotu))

			time.Sleep(time.Millisecond)

			require.NoError(t, m.Clear(time.Nanosecond, false))
			got, err = m.Read(id)
			require.Error(t, err)
		})
	}
}
