package link

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ory/x/configx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/x/logrusx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/stringslice"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
)

func TestRecoveryToken(t *testing.T) {
	conf, err := config.New(context.Background(), logrusx.New("", ""), configx.SkipValidation())
	require.NoError(t, err)
	req := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
	t.Run("func=NewSelfServiceRecoveryToken", func(t *testing.T) {
		t.Run("case=creates unique tokens", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			tokens := make([]string, 10)
			for k := range tokens {
				tokens[k] = NewSelfServiceRecoveryToken(nil, f, time.Hour).Token
			}

			assert.Len(t, stringslice.Unique(tokens), len(tokens))
		})
	})
	t.Run("method=Valid", func(t *testing.T) {
		t.Run("case=is invalid when the flow is expired", func(t *testing.T) {
			f, err := recovery.NewFlow(conf, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			token := NewSelfServiceRecoveryToken(nil, f, -time.Hour)
			require.Error(t, token.Valid())
			assert.EqualError(t, token.Valid(), f.Valid().Error())
		})
	})
}
