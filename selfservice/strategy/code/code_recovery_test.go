// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/x"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
)

func TestRecoveryCode(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)

	newCode := func(expiresIn time.Duration, f *recovery.Flow) *code.RecoveryCode {
		return &code.RecoveryCode{
			ID:        x.NewUUID(),
			FlowID:    f.ID,
			ExpiresAt: time.Now().Add(expiresIn),
		}
	}

	req := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}

	t.Run("method=IsExpired", func(t *testing.T) {
		t.Run("case=returns true if flow is expired", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(-time.Hour, f)
			require.True(t, c.IsExpired())
		})
		t.Run("case=returns false if flow is not expired", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(time.Hour, f)
			require.False(t, c.IsExpired())
		})
	})

	t.Run("method=WasUsed", func(t *testing.T) {
		t.Run("case=returns true if flow has been used", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(time.Hour, f)
			c.UsedAt = sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			}
			require.True(t, c.WasUsed())
		})
		t.Run("case=returns false if flow has not been used", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(time.Hour, f)
			c.UsedAt = sql.NullTime{
				Valid: false,
			}
			require.False(t, c.WasUsed())
		})
	})
}

func TestRecoveryCodeType(t *testing.T) {
	assert.Equal(t, 1, int(code.RecoveryCodeTypeAdmin))
	assert.Equal(t, 2, int(code.RecoveryCodeTypeSelfService))
}
