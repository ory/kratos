// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/sessiontokenexchange"
	"github.com/ory/x/randx"
)

type testParams struct {
	flowID, sessionID      uuid.UUID
	initCode, returnToCode string
}

func newParams() testParams {
	return testParams{
		flowID:       uuid.Must(uuid.NewV4()),
		sessionID:    uuid.Must(uuid.NewV4()),
		initCode:     randx.MustString(64, randx.AlphaNum),
		returnToCode: randx.MustString(64, randx.AlphaNum),
	}
}
func (t *testParams) setCodes(e *sessiontokenexchange.Exchanger) {
	t.initCode = e.InitCode
	t.returnToCode = e.ReturnToCode
}

func TestPersister(ctx context.Context, _ *config.Config, p interface {
	persistence.Persister
}) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		nid, p := testhelpers.NewNetworkUnlessExisting(t, ctx, p)

		t.Run("suite=create-update-get", func(t *testing.T) {
			t.Parallel()
			params := newParams()

			t.Run("step=create", func(t *testing.T) {
				e, err := p.CreateSessionTokenExchanger(ctx, params.flowID)
				require.NoError(t, err)
				params.setCodes(e)
				codes, ok, err := p.CodeForFlow(ctx, params.flowID)
				assert.True(t, ok)
				assert.NoError(t, err)
				assert.Equal(t, params.initCode, codes.InitCode)
				assert.Equal(t, params.returnToCode, codes.ReturnToCode)
			})
			t.Run("step=update", func(t *testing.T) {
				require.NoError(t, p.UpdateSessionOnExchanger(ctx, params.flowID, params.sessionID))
			})
			t.Run("step=get", func(t *testing.T) {
				e, err := p.GetExchangerFromCode(ctx, params.initCode, params.returnToCode)
				require.NoError(t, err)

				assert.Equal(t, params.sessionID, e.SessionID.UUID)
				assert.Equal(t, nid, e.NID)
			})
		})

		t.Run("suite=CodeForFlow", func(t *testing.T) {
			t.Parallel()

			t.Run("case=returns false for non-existing flow", func(t *testing.T) {
				t.Parallel()
				_, ok, err := p.CodeForFlow(ctx, uuid.Must(uuid.NewV4()))
				assert.False(t, ok)
				assert.NoError(t, err)
			})
		})

		t.Run("suite=MoveToNewFlow", func(t *testing.T) {
			t.Parallel()

			t.Run("case=move to new flow", func(t *testing.T) {
				params := newParams()
				other := newParams()

				e, err := p.CreateSessionTokenExchanger(ctx, params.flowID)
				require.NoError(t, err)
				params.setCodes(e)
				require.NoError(t, p.MoveToNewFlow(ctx, params.flowID, other.flowID))
				require.NoError(t, p.UpdateSessionOnExchanger(ctx, other.flowID, params.sessionID))

				e, err = p.GetExchangerFromCode(ctx, params.initCode, params.returnToCode)
				require.NoError(t, err)
				assert.Equal(t, params.sessionID, e.SessionID.UUID)
			})
		})

		t.Run("suite=GetExchangerFromCode", func(t *testing.T) {
			t.Parallel()

			t.Run("case=errors if session not found", func(t *testing.T) {
				t.Parallel()
				params := newParams()

				e, err := p.CreateSessionTokenExchanger(ctx, params.flowID)
				require.NoError(t, err)
				params.setCodes(e)

				e, err = p.GetExchangerFromCode(ctx, params.initCode, params.returnToCode)

				assert.Error(t, err)
				assert.Nil(t, e)
			})

			t.Run("case=errors if code is invalid", func(t *testing.T) {
				t.Parallel()
				params := newParams()
				other := newParams()

				e, err := p.CreateSessionTokenExchanger(ctx, params.flowID)
				require.NoError(t, err)
				params.setCodes(e)

				require.NoError(t, p.UpdateSessionOnExchanger(ctx, params.flowID, params.sessionID))
				e, err = p.GetExchangerFromCode(ctx, other.initCode, other.returnToCode)

				assert.Error(t, err)
				assert.Nil(t, e)
			})

			t.Run("case=errors if code is empty", func(t *testing.T) {
				t.Parallel()
				params := newParams()

				e, err := p.CreateSessionTokenExchanger(ctx, params.flowID)
				require.NoError(t, err)
				params.setCodes(e)

				require.NoError(t, p.UpdateSessionOnExchanger(ctx, params.flowID, params.sessionID))
				e, err = p.GetExchangerFromCode(ctx, "", "")

				assert.Error(t, err)
				assert.Nil(t, e)
			})

			t.Run("case=errors if other network ID", func(t *testing.T) {
				t.Parallel()
				params := newParams()
				otherNID := uuid.Must(uuid.NewV4())

				e, err := p.CreateSessionTokenExchanger(ctx, params.flowID)
				require.NoError(t, err)
				params.setCodes(e)

				require.NoError(t, p.UpdateSessionOnExchanger(ctx, params.flowID, params.sessionID))
				e, err = p.WithNetworkID(otherNID).GetExchangerFromCode(ctx, params.initCode, params.returnToCode)

				assert.Error(t, err)
				assert.Nil(t, e)
			})
		})
	}
}
