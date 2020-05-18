package continuity

import (
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/x"
)

func TestContainer(t *testing.T) {
	id := x.NewUUID()
	for k, tc := range []struct {
		c    *Container
		i    uuid.UUID
		pass bool
	}{
		{
			c: &Container{
				ExpiresAt: time.Now().Add(-time.Minute),
			},
			pass: false,
		},
		{
			c: &Container{
				ExpiresAt: time.Now().Add(-time.Minute),
			},
			i:    x.NewUUID(),
			pass: false,
		},
		{
			c: &Container{
				IdentityID: x.PointToUUID(x.NewUUID()),
				ExpiresAt:  time.Now().Add(-time.Minute),
			},
			i:    x.NewUUID(),
			pass: false,
		},
		{
			c: &Container{
				ExpiresAt: time.Now().Add(time.Minute),
			},
			pass: true,
		},
		{
			c: &Container{
				IdentityID: x.PointToUUID(id),
				ExpiresAt:  time.Now().Add(time.Minute),
			},
			i:    id,
			pass: true,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			err := tc.c.Valid(tc.i)
			if tc.pass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
