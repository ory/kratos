// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package hook_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

func TestVerifier(t *testing.T) {
	ctx := context.Background()
	u := &http.Request{URL: urlx.ParseOrPanic("https://www.ory.sh/")}
	for k, hf := range map[string]func(*hook.Verifier, *identity.Identity, flow.Flow) error{
		"settings": func(h *hook.Verifier, i *identity.Identity, f flow.Flow) error {
			return h.ExecuteSettingsPostPersistHook(
				httptest.NewRecorder(), u, f.(*settings.Flow), i)
		},
		"register": func(h *hook.Verifier, i *identity.Identity, f flow.Flow) error {
			return h.ExecutePostRegistrationPostPersistHook(
				httptest.NewRecorder(), u, f.(*registration.Flow), &session.Session{ID: x.NewUUID(), Identity: i})
		},
	} {
		t.Run("name="+k, func(t *testing.T) {
			conf, reg := internal.NewFastRegistryWithMocks(t)
			testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/verify.schema.json")
			conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "https://www.ory.sh/")
			conf.MustSet(ctx, config.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")

			i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			i.Traits = identity.Traits(`{"emails":["foo@ory.sh","bar@ory.sh","baz@ory.sh"]}`)
			require.NoError(t, reg.IdentityManager().Create(context.Background(), i))

			actual, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, "foo@ory.sh")
			require.NoError(t, err)
			assert.EqualValues(t, "foo@ory.sh", actual.Value)

			actual, err = reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, "bar@ory.sh")
			require.NoError(t, err)
			assert.EqualValues(t, "bar@ory.sh", actual.Value)

			actual, err = reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, "baz@ory.sh")
			require.NoError(t, err)
			assert.EqualValues(t, "baz@ory.sh", actual.Value)

			verifiedAt := sqlxx.NullTime(time.Now())
			actual.Status = identity.VerifiableAddressStatusCompleted
			actual.Verified = true
			actual.VerifiedAt = &verifiedAt
			require.NoError(t, reg.PrivilegedIdentityPool().UpdateVerifiableAddress(context.Background(), actual))

			i, err = reg.IdentityPool().GetIdentity(context.Background(), i.ID, identity.ExpandDefault)
			require.NoError(t, err)

			var originalFlow flow.FlowWithContinueWith
			switch k {
			case "settings":
				originalFlow = &settings.Flow{RequestURL: "http://foo.com/settings?after_verification_return_to=verification_callback"}
			case "register":
				originalFlow = &registration.Flow{RequestURL: "http://foo.com/registration?after_verification_return_to=verification_callback"}
			default:
				t.FailNow()
			}

			h := hook.NewVerifier(reg)
			require.NoError(t, hf(h, i, originalFlow))
			assert.Lenf(t, originalFlow.ContinueWith(), 2, "%#ßv", originalFlow.ContinueWith())
			assertContinueWithAddresses(t, originalFlow.ContinueWith(), []string{"foo@ory.sh", "bar@ory.sh"})
			vf := originalFlow.ContinueWith()[0]
			assert.IsType(t, &flow.ContinueWithVerificationUI{}, vf)
			fView := vf.(*flow.ContinueWithVerificationUI).Flow

			expectedVerificationFlow, err := reg.VerificationFlowPersister().GetVerificationFlow(ctx, fView.ID)
			require.NoError(t, err)
			require.Equal(t, expectedVerificationFlow.State, flow.StateEmailSent)

			messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
			require.NoError(t, err)
			require.Len(t, messages, 2)

			recipients := make([]string, len(messages))
			for k, m := range messages {
				recipients[k] = m.Recipient
			}

			assert.Contains(t, recipients, "foo@ory.sh")
			assert.Contains(t, recipients, "bar@ory.sh")
			// Email to baz@ory.sh is skipped because it is verified already.
			assert.NotContains(t, recipients, "baz@ory.sh")

			// these addresses will be marked as sent and won't be sent again by the settings hook
			address1, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, "foo@ory.sh")
			require.NoError(t, err)
			assert.EqualValues(t, identity.VerifiableAddressStatusSent, address1.Status)
			address2, err := reg.IdentityPool().FindVerifiableAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, "bar@ory.sh")
			require.NoError(t, err)
			assert.EqualValues(t, identity.VerifiableAddressStatusSent, address2.Status)

			switch k {
			case "settings":
				originalFlow = &settings.Flow{RequestURL: "http://foo.com/settings?after_verification_return_to=verification_callback"}
			case "register":
				originalFlow = &registration.Flow{RequestURL: "http://foo.com/registration?after_verification_return_to=verification_callback"}
			default:
				t.FailNow()
			}
			require.NoError(t, hf(h, i, originalFlow))

			assert.Emptyf(t, originalFlow.ContinueWith(), "%+v", originalFlow.ContinueWith())

			require.NoError(t, err)
			messages, err = reg.CourierPersister().NextMessages(context.Background(), 12)
			require.EqualError(t, err, courier.ErrQueueEmpty.Error())
			assert.Len(t, messages, 0)
		})
	}
}

func assertContinueWithAddresses(t *testing.T, cs []flow.ContinueWith, addresses []string) {
	t.Helper()

	require.Len(t, cs, len(addresses))

	e := make([]string, len(cs))

	for i, c := range cs {
		require.IsType(t, &flow.ContinueWithVerificationUI{}, c)
		fView := c.(*flow.ContinueWithVerificationUI).Flow
		e[i] = fView.VerifiableAddress
	}

	require.ElementsMatch(t, addresses, e)
}
