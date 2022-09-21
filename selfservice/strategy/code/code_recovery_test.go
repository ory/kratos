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

	"github.com/ory/x/stringslice"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
)

func TestRecoveryCode(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)

	req := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
	t.Run("func=NewSelfServiceRecoveryCode", func(t *testing.T) {
		t.Run("case=creates unique tokens", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			codes := make([]string, 10)
			for k := range codes {
				codes[k] = code.NewSelfServiceRecoveryCode(x.NewUUID(), nil, f, time.Hour).Code
			}

			assert.Len(t, stringslice.Unique(codes), len(codes))
		})
	})

	t.Run("func=NewAdminRecoveryCode", func(t *testing.T) {
		t.Run("case=creates unique tokens", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			codes := make([]string, 10)
			for k := range codes {
				codes[k] = code.NewAdminRecoveryCode(x.NewUUID(), f.ID, time.Hour).Code
			}

			assert.Len(t, stringslice.Unique(codes), len(codes))
		})
	})

	t.Run("method=IsExpired", func(t *testing.T) {
		t.Run("case=returns true if flow is expired", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			token := code.NewSelfServiceRecoveryCode(x.NewUUID(), nil, f, -time.Hour)
			require.True(t, token.IsExpired())
		})
		t.Run("case=returns false if flow is not expired", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			token := code.NewSelfServiceRecoveryCode(x.NewUUID(), nil, f, time.Hour)
			require.False(t, token.IsExpired())
		})
	})

	t.Run("method=WasUsed", func(t *testing.T) {
		t.Run("case=returns true if flow has been used", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			token := code.NewSelfServiceRecoveryCode(x.NewUUID(), nil, f, -time.Hour)
			token.UsedAt = sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			}
			require.True(t, token.WasUsed())
		})
		t.Run("case=returns false if flow has not been used", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			token := code.NewSelfServiceRecoveryCode(x.NewUUID(), nil, f, -time.Hour)
			token.UsedAt = sql.NullTime{
				Valid: false,
			}
			require.False(t, token.WasUsed())
		})
	})
}

func TestRecoveryCodeType(t *testing.T) {
	assert.Equal(t, 1, int(code.RecoveryCodeTypeAdmin))
	assert.Equal(t, 2, int(code.RecoveryCodeTypeSelfService))
}
