package token_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/token"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/urlx"
)

func TestRecoveryToken(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)

	req := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
	t.Run("func=NewLinkRecovery", func(t *testing.T) {
		t.Run("case=creates unique tokens", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			tokens := make([]string, 10)
			for k := range tokens {
				tokens[k] = token.NewLinkRecovery(nil, f, time.Hour).Token
			}

			assert.Len(t, stringslice.Unique(tokens), len(tokens))
		})
	})
	t.Run("method=Valid", func(t *testing.T) {
		t.Run("case=is invalid when the flow is expired", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			tkn := token.NewLinkRecovery(nil, f, -time.Hour)
			require.Error(t, tkn.Valid())
			assert.EqualError(t, tkn.Valid(), f.Valid().Error())
		})
	})
}

func TestRecoveryTokenType(t *testing.T) {
	assert.Equal(t, 1, int(link.RecoveryTokenTypeAdmin))
	assert.Equal(t, 2, int(link.RecoveryTokenTypeSelfService))
}
