package profile_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/profile"
	"github.com/ory/kratos/session"
)

type mockPostHook struct {
	err error
}

func (m *mockPostHook) ExecuteProfileManagementPostHook(w http.ResponseWriter, r *http.Request, a *profile.Request, s *session.Session) error {
	return m.err
}

type loginExecutorDependenciesMock struct {
	preErr []error
}

func (m *loginExecutorDependenciesMock) PostLoginHooks(credentialsType identity.CredentialsType) []profile.PostHookExecutor {
	hooks := make([]profile.PostHookExecutor, len(m.preErr))
	for k := range hooks {
		hooks[k] = &mockPostHook{m.preErr[k]}
	}
	return hooks
}

func TestProfileManagementExecutor(t *testing.T) {
	t.Run("method=PostProfileManagementHook", func(t *testing.T) {
		for k, tc := range []struct {
			hooks          []profile.PostHookExecutor
			expectSchemaID string
			expectErr      error
		}{
			{hooks: nil},
			{hooks: []profile.PostHookExecutor{}},
			{hooks: []profile.PostHookExecutor{&mockPostHook{err: errors.New("err")}}, expectErr: errors.New("err")},
			{hooks: []profile.PostHookExecutor{
				new(mockPostHook),
				&mockPostHook{err: errors.New("err")}}, expectErr: errors.New("err"),
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				conf, reg := internal.NewRegistryDefault(t)

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))
				i.TraitsSchemaID = ""
				i.Traits = identity.Traits(`{}`)
				viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/identity.schema.json")
				viper.Set(configuration.ViperKeyURLsSelfPublic, "http://example.com")
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.TODO(), &i))

				sess := &session.Session{Identity: &i, IdentityID: i.ID}
				pr := profile.NewRequest(time.Hour, &http.Request{URL: new(url.URL)}, sess)
				require.NoError(t, reg.ProfileRequestPersister().CreateProfileRequest(context.Background(), pr))

				e := profile.NewHookExecutor(reg, conf)
				err := e.PostProfileManagementHook(nil, &http.Request{}, tc.hooks, pr, sess, &i)
				if tc.expectErr != nil {
					require.EqualError(t, err, tc.expectErr.Error())
					return
				}

				require.NoError(t, err)
				if tc.expectSchemaID != "" {
					got, err := reg.IdentityPool().GetIdentity(context.TODO(), i.ID)
					require.NoError(t, err)
					assert.EqualValues(t, tc.expectSchemaID, got.TraitsSchemaID)
				}
			})
		}
	})
}
