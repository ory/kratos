// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/x"
	"github.com/ory/x/urlx"
)

func TestRecoveryCode(t *testing.T) {
	_, reg := pkg.NewVeryFastRegistryWithoutDB(t)

	newCode := func(expiresIn time.Duration, f *recovery.Flow) *code.RecoveryCode {
		return &code.RecoveryCode{
			ID:        x.NewUUID(),
			FlowID:    f.ID,
			ExpiresAt: time.Now().Add(expiresIn),
		}
	}

	req := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.com/")}
	t.Run("method=Validate", func(t *testing.T) {
		t.Parallel()

		t.Run("case=returns ErrCodeNotFound if code is expired", func(t *testing.T) {
			f, err := recovery.NewFlow(reg, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(-time.Hour, f)
			require.ErrorIs(t, c.Validate(reg.Clock()), code.ErrCodeNotFound())
		})
		t.Run("case=expired code does not return flow.ExpiredError", func(t *testing.T) {
			f, err := recovery.NewFlow(reg, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(-time.Hour, f)
			expired := new(flow.ExpiredError)
			require.False(t, errors.As(c.Validate(reg.Clock()), &expired), "expired code should not return flow.ExpiredError")
		})
		t.Run("case=returns no error if flow is not expired", func(t *testing.T) {
			f, err := recovery.NewFlow(reg, time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(time.Hour, f)
			require.NoError(t, c.Validate(reg.Clock()))
		})

		t.Run("case=returns error if flow has been used", func(t *testing.T) {
			f, err := recovery.NewFlow(reg, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(time.Hour, f)
			c.UsedAt = sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			}
			require.ErrorIs(t, c.Validate(reg.Clock()), code.ErrCodeAlreadyUsed())
		})

		t.Run("case=returns no error if flow has not been used", func(t *testing.T) {
			f, err := recovery.NewFlow(reg, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(time.Hour, f)
			c.UsedAt = sql.NullTime{
				Valid: false,
			}
			require.NoError(t, c.Validate(reg.Clock()))
		})

		t.Run("case=returns error if flow is nil", func(t *testing.T) {
			var c *code.RecoveryCode
			require.ErrorIs(t, c.Validate(reg.Clock()), code.ErrCodeNotFound())
		})
	})
}

func TestRecoveryCodeType(t *testing.T) {
	assert.Equal(t, 1, int(code.RecoveryCodeTypeAdmin))
	assert.Equal(t, 2, int(code.RecoveryCodeTypeSelfService))
}
