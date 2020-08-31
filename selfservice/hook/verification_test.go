package hook_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/sqlxx"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/hook"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

func TestVerifier(t *testing.T) {
	for k, hf := range map[string]func(*hook.Verifier, *identity.Identity) error{
		"settings": func(h *hook.Verifier, i *identity.Identity) error {
			return h.ExecuteSettingsPostPersistHook(
				httptest.NewRecorder(), new(http.Request), nil, i)
		},
		"register": func(h *hook.Verifier, i *identity.Identity) error {
			return h.ExecutePostRegistrationPostPersistHook(
				httptest.NewRecorder(), new(http.Request), nil, &session.Session{ID: x.NewUUID(), Identity: i})
		},
	} {
		t.Run("name="+k, func(t *testing.T) {
			conf, reg := internal.NewFastRegistryWithMocks(t)
			viper.Set(configuration.ViperKeyDefaultIdentitySchemaURL, "file://./stub/verify.schema.json")
			viper.Set(configuration.ViperKeyPublicBaseURL, "https://www.ory.sh/")
			viper.Set(configuration.ViperKeyCourierSMTPURL, "smtp://foo@bar@dev.null/")

			i := identity.NewIdentity(configuration.DefaultIdentityTraitsSchemaID)
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
			actual.Status = identity.VerifiableAddressStatusCompleted
			actual.Verified = true
			actual.VerifiedAt = sqlxx.NullTime(time.Now())
			require.NoError(t, reg.PrivilegedIdentityPool().UpdateVerifiableAddress(context.Background(), actual))

			i, err = reg.IdentityPool().GetIdentity(context.Background(), i.ID)
			require.NoError(t, err)

			h := hook.NewVerifier(reg, conf)
			require.NoError(t, hf(h, i))

			messages, err := reg.CourierPersister().NextMessages(context.Background(), 12)
			require.NoError(t, err)
			require.Len(t, messages, 2)

			assert.EqualValues(t, "foo@ory.sh", messages[0].Recipient)
			assert.EqualValues(t, "bar@ory.sh", messages[1].Recipient)
			// Email to baz@ory.sh is skipped because it is verified already.
		})
	}
}
