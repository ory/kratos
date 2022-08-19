package code_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/strategy/code"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/stringslice"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
)

func TestRecoveryToken(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)

	req := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
	t.Run("func=NewSelfServiceRecoveryCode", func(t *testing.T) {
		t.Run("case=creates unique tokens", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			codes := make([]string, 10)
			for k := range codes {
				codes[k] = code.NewSelfServiceRecoveryCode(nil, f, time.Hour).Code
			}

			assert.Len(t, stringslice.Unique(codes), len(codes))
		})
	})
	t.Run("method=Valid", func(t *testing.T) {
		t.Run("case=is invalid when the flow is expired", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			token := code.NewSelfServiceRecoveryCode(nil, f, -time.Hour)
			require.Error(t, token.Valid())
			assert.EqualError(t, token.Valid(), f.Valid().Error())
		})
	})
}

func TestRecoveryCodeType(t *testing.T) {
	assert.Equal(t, 1, int(code.RecoveryCodeTypeAdmin))
	assert.Equal(t, 2, int(code.RecoveryCodeTypeSelfService))
}
