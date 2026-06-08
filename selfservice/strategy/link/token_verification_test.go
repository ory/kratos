// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package link_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/ory/kratos/pkg"
	"github.com/ory/kratos/selfservice/strategy/link"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/clock"
	"github.com/ory/x/stringslice"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/verification"
)

func TestVerificationToken(t *testing.T) {
	_, reg := pkg.NewFastRegistryWithMocks(t)

	req := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.com/")}
	t.Run("func=NewSelfServiceVerificationToken", func(t *testing.T) {
		t.Run("case=creates unique tokens", func(t *testing.T) {
			f, err := verification.NewFlow(reg, time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			tokens := make([]string, 10)
			for k := range tokens {
				tokens[k] = link.NewSelfServiceVerificationToken(nil, f, time.Hour).Token
			}

			assert.Len(t, stringslice.Unique(tokens), len(tokens))
		})
	})
	t.Run("method=Valid", func(t *testing.T) {
		t.Run("case=is invalid when the flow is expired", func(t *testing.T) {
			f, err := verification.NewFlow(reg, -time.Hour, "", req, nil, flow.TypeBrowser)
			require.NoError(t, err)

			token := link.NewSelfServiceVerificationToken(nil, f, -time.Hour)
			require.Error(t, token.Valid(clock.New()))
			assert.EqualError(t, token.Valid(clock.New()), f.Valid(clock.New()).Error())
		})
	})
}
