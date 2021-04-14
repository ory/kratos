package errorx

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/kratos/persistence"
	"github.com/ory/x/sqlcon"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"

	"github.com/ory/kratos/x"
)

func TestPersister(ctx context.Context, p persistence.Persister) func(t *testing.T) {
	toJSON := func(t *testing.T, in interface{}) string {
		out, err := json.Marshal(in)
		require.NoError(t, err)
		return string(out)
	}

	return func(t *testing.T) {
		_, p := testhelpers.NewNetworkUnlessExisting(t, ctx, p)

		t.Run("case=not found", func(t *testing.T) {
			_, err := p.Read(ctx, x.NewUUID())
			require.Error(t, err)
		})

		t.Run("case=en- and decode properly", func(t *testing.T) {
			actualID, err := p.Add(ctx, "nosurf", herodot.ErrNotFound.WithReason("foobar"))
			require.NoError(t, err)

			actual, err := p.Read(ctx, actualID)
			require.NoError(t, err)

			assert.JSONEq(t, `{"code":404,"status":"Not Found","reason":"foobar","message":"The requested resource could not be found"}`, gjson.Get(toJSON(t, actual), "errors.0").String(), toJSON(t, actual))
		})

		t.Run("case=clear", func(t *testing.T) {
			actualID, err := p.Add(ctx, "nosurf", herodot.ErrNotFound.WithReason("foobar"))
			require.NoError(t, err)

			_, err = p.Read(ctx, actualID)
			require.NoError(t, err)

			// We need to wait for at least one second or MySQL will randomly fail as it does not support
			// millisecond resolution on timestamp columns.
			time.Sleep(time.Second + time.Millisecond*500)
			require.NoError(t, p.Clear(ctx, time.Second, false))
			got, err := p.Read(ctx, actualID)
			require.Error(t, err, "%+v", got)
		})

		t.Run("case=network", func(t *testing.T) {
			t.Run("can not read error from another network", func(t *testing.T) {
				created, err := p.Add(ctx, "nosurf", herodot.ErrNotFound.WithReason("foobar"))
				require.NoError(t, err)

				time.Sleep(time.Second + time.Millisecond*500)

				_, other := testhelpers.NewNetwork(t, ctx, p)
				_, err = other.Read(ctx, created)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)

				t.Run("can not clear another network", func(t *testing.T) {
					_, other := testhelpers.NewNetwork(t, ctx, p)
					require.NoError(t, other.Clear(ctx, time.Second, true))

					c, err := p.Read(ctx, created)
					require.NoError(t, err)
					assert.Contains(t, string(c.Errors), "foobar")
				})
			})
		})
	}
}
