// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/x"
	"github.com/ory/x/urlx"
)

func TestRegistrationCode(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)
	newCode := func(expiresIn time.Duration, f *registration.Flow) *code.RegistrationCode {
		return &code.RegistrationCode{
			ID:        x.NewUUID(),
			FlowID:    f.ID,
			ExpiresAt: time.Now().Add(expiresIn),
		}
	}

	req := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
	t.Run("method=Validate", func(t *testing.T) {
		t.Parallel()

		t.Run("case=returns error if flow is expired", func(t *testing.T) {
			f, err := registration.NewFlow(conf, -time.Hour, "", req, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(-time.Hour, f)
			expected := new(flow.ExpiredError)
			require.ErrorAs(t, c.Validate(), &expected)
		})
		t.Run("case=returns no error if flow is not expired", func(t *testing.T) {
			f, err := registration.NewFlow(conf, time.Hour, "", req, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(time.Hour, f)
			require.NoError(t, c.Validate())
		})

		t.Run("case=returns error if flow has been used", func(t *testing.T) {
			f, err := registration.NewFlow(conf, -time.Hour, "", req, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(time.Hour, f)
			c.UsedAt = sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			}
			require.ErrorIs(t, c.Validate(), code.ErrCodeAlreadyUsed)
		})

		t.Run("case=returns no error if flow has not been used", func(t *testing.T) {
			f, err := registration.NewFlow(conf, -time.Hour, "", req, flow.TypeBrowser)
			require.NoError(t, err)

			c := newCode(time.Hour, f)
			c.UsedAt = sql.NullTime{
				Valid: false,
			}
			require.NoError(t, c.Validate())
		})

		t.Run("case=returns error if flow is nil", func(t *testing.T) {
			var c *code.RegistrationCode
			require.ErrorIs(t, c.Validate(), code.ErrCodeNotFound)
		})
	})
}
