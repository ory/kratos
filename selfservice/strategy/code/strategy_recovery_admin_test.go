// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package code_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	kratos "github.com/ory/kratos/internal/httpclient"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/recovery"
	"github.com/ory/kratos/selfservice/strategy/code"
	"github.com/ory/kratos/x"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/pointerx"
	"github.com/ory/x/snapshotx"
)

func TestAdminStrategy(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	initViper(t, ctx, conf)

	_ = testhelpers.NewRecoveryUIFlowEchoServer(t, reg)
	_ = testhelpers.NewSettingsUIFlowEchoServer(t, reg)
	_ = testhelpers.NewLoginUIFlowEchoServer(t, reg)
	_ = testhelpers.NewErrorTestServer(t, reg)

	publicTS, adminTS := testhelpers.NewKratosServer(t, reg)
	adminSDK := testhelpers.NewSDKClient(adminTS)

	createCode := func(id string, expiresIn *string) (*kratos.RecoveryCodeForIdentity, *http.Response, error) {
		return adminSDK.IdentityApi.
			CreateRecoveryCodeForIdentity(context.Background()).
			CreateRecoveryCodeForIdentityBody(
				kratos.CreateRecoveryCodeForIdentityBody{
					IdentityId: id,
					ExpiresIn:  expiresIn,
				}).Execute()
	}

	t.Run("no panic on empty body #1384", func(t *testing.T) {
		ctx := context.Background()
		s, err := reg.RecoveryStrategies(ctx).Strategy("code")
		require.NoError(t, err)
		w := httptest.NewRecorder()
		r := &http.Request{URL: new(url.URL)}
		f, err := recovery.NewFlow(reg.Config(), time.Minute, "", r, s, flow.TypeBrowser)
		require.NoError(t, err)
		require.NotPanics(t, func() {
			require.Error(t, s.(*code.Strategy).HandleRecoveryError(w, r, f, nil, errors.New("test")))
		})
	})

	t.Run("description=should not be able to recover an account that does not exist", func(t *testing.T) {
		_, _, err := createCode(x.NewUUID().String(), nil)

		require.IsType(t, err, new(kratos.GenericOpenAPIError), "%T", err)
		snapshotx.SnapshotT(t, err.(*kratos.GenericOpenAPIError).Model())
	})

	t.Run("description=should fail on malformed expiry time", func(t *testing.T) {
		_, _, err := createCode(x.NewUUID().String(), pointerx.String("not-a-valid-value"))
		require.IsType(t, err, new(kratos.GenericOpenAPIError), "%T", err)
		snapshotx.SnapshotT(t, err.(*kratos.GenericOpenAPIError).Model())
	})

	t.Run("description=should fail on negative expiry time", func(t *testing.T) {
		_, _, err := createCode(x.NewUUID().String(), pointerx.String("-1h"))
		require.IsType(t, err, new(kratos.GenericOpenAPIError), "%T", err)
		snapshotx.SnapshotT(t, err.(*kratos.GenericOpenAPIError).Model())
	})

	submitRecoveryLink := func(t *testing.T, link string, code string) []byte {
		t.Helper()
		res, err := publicTS.Client().Get(link)
		require.NoError(t, err)
		body := ioutilx.MustReadAll(res.Body)

		action := gjson.GetBytes(body, "ui.action").String()
		require.NotEmpty(t, action)

		res, err = publicTS.Client().PostForm(action, url.Values{
			"code": {code},
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		return ioutilx.MustReadAll(res.Body)
	}

	t.Run("description=should create code without email", func(t *testing.T) {
		id := identity.Identity{Traits: identity.Traits(`{}`)}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		code, _, err := createCode(id.ID.String(), nil)
		require.NoError(t, err)

		require.NotEmpty(t, code.RecoveryLink)
		require.Contains(t, code.RecoveryLink, "flow=")
		require.NotContains(t, code.RecoveryLink, "code=")
		require.NotEmpty(t, code.RecoveryCode)
		require.True(t, code.ExpiresAt.Before(time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan(ctx))))

		body := submitRecoveryLink(t, code.RecoveryLink, code.RecoveryCode)
		testhelpers.AssertMessage(t, body, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.")
	})

	t.Run("description=should not be able to recover with expired code", func(t *testing.T) {
		recoveryEmail := "recover.expired@ory.sh"
		id := identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"email":"%s"}`, recoveryEmail))}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		code, _, err := createCode(id.ID.String(), pointerx.String("100ms"))
		require.NoError(t, err)

		time.Sleep(time.Millisecond * 100)
		require.NotEmpty(t, code.RecoveryLink)
		require.True(t, code.ExpiresAt.Before(time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan(ctx))))

		body := submitRecoveryLink(t, code.RecoveryLink, code.RecoveryCode)
		testhelpers.AssertMessage(t, body, "The recovery flow expired 0.00 minutes ago, please try again.")

		// The recovery address should not be verified if the flow was initiated by the admins
		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
		assert.NoError(t, err)
		assert.False(t, addr.Verified)
		assert.Nil(t, addr.VerifiedAt)
		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	})

	t.Run("description=should create a valid recovery link and set the expiry time as well and recover the account", func(t *testing.T) {
		recoveryEmail := "recoverme@ory.sh"
		id := identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"email":"%s"}`, recoveryEmail))}

		require.NoError(t, reg.IdentityManager().Create(context.Background(),
			&id, identity.ManagerAllowWriteProtectedTraits))

		code, _, err := createCode(id.ID.String(), nil)
		require.NoError(t, err)

		require.NotEmpty(t, code.RecoveryLink)
		require.True(t, code.ExpiresAt.Before(time.Now().Add(conf.SelfServiceFlowRecoveryRequestLifespan(ctx)+time.Second)))

		body := submitRecoveryLink(t, code.RecoveryLink, code.RecoveryCode)

		testhelpers.AssertMessage(t, body, "You successfully recovered your account. Please change your password or set up an alternative login method (e.g. social sign in) within the next 60.00 minutes.")

		addr, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, recoveryEmail)
		assert.NoError(t, err)
		assert.False(t, addr.Verified)
		assert.Nil(t, addr.VerifiedAt)
		assert.Equal(t, identity.VerifiableAddressStatusPending, addr.Status)
	})

	t.Run("case=should not be able to use code from different flow", func(t *testing.T) {
		email := testhelpers.RandomEmail()
		i := createIdentityToRecover(t, reg, email)

		c1, _, err := createCode(i.ID.String(), pointerx.String("1h"))
		require.NoError(t, err)
		c2, _, err := createCode(i.ID.String(), pointerx.String("1h"))
		require.NoError(t, err)
		code2 := c2.RecoveryCode
		require.NotEmpty(t, code2)

		body := submitRecoveryLink(t, c1.RecoveryLink, c2.RecoveryCode)

		testhelpers.AssertMessage(t, body, "The recovery code is invalid or has already been used. Please try again.")
	})

	t.Run("case=form should not contain email field when creating recovery code", func(t *testing.T) {
		email := testhelpers.RandomEmail()
		i := createIdentityToRecover(t, reg, email)

		c1, _, err := createCode(i.ID.String(), pointerx.String("1h"))
		require.NoError(t, err)

		res, err := http.Get(c1.RecoveryLink)
		require.NoError(t, err)
		body := ioutilx.MustReadAll(res.Body)

		snapshotx.SnapshotT(t, json.RawMessage(gjson.GetBytes(body, "ui.nodes").String()))
	})
}
