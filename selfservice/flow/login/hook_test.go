package login_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/session"
)

type loginPreHookMock struct {
	err error
}

func (m *loginPreHookMock) ExecuteLoginPreHook(w http.ResponseWriter, r *http.Request, a *login.Request) error {
	return m.err
}

type mockPostHook struct {
	err            error
	modifyIdentity bool
}

var updatedSchema = configuration.SchemaConfig{
	ID:  "updatedSchema",
	URL: "file://./stub/updated.schema.json",
}

func (m *mockPostHook) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, a *login.Request, s *session.Session) error {
	if m.modifyIdentity {
		i := s.Identity
		i.TraitsSchemaID = updatedSchema.ID
		s.UpdateIdentity(i)
	}
	return m.err
}

type loginExecutorDependenciesMock struct {
	preErr []error
}

func (m *loginExecutorDependenciesMock) PostLoginHooks(credentialsType identity.CredentialsType) []login.PostHookExecutor {
	return []login.PostHookExecutor{}
}

func (m *loginExecutorDependenciesMock) PrivilegedIdentityPool() identity.PrivilegedPool {
	return nil
}

func (m *loginExecutorDependenciesMock) PreLoginHooks() []login.PreHookExecutor {
	hooks := make([]login.PreHookExecutor, len(m.preErr))
	for k := range hooks {
		hooks[k] = &loginPreHookMock{m.preErr[k]}
	}
	return hooks
}

func TestLoginExecutor(t *testing.T) {
	t.Run("method=PostLoginHook", func(t *testing.T) {
		for k, tc := range []struct {
			hooks          []login.PostHookExecutor
			expectSchemaID string
			expectErr      error
		}{
			{hooks: nil},
			{hooks: []login.PostHookExecutor{}},
			{hooks: []login.PostHookExecutor{&mockPostHook{err: errors.New("err")}}, expectErr: errors.New("err")},
			{hooks: []login.PostHookExecutor{
				new(mockPostHook),
				&mockPostHook{err: errors.New("err")}}, expectErr: errors.New("err"),
			},
			{
				hooks: []login.PostHookExecutor{
					new(mockPostHook),
					&mockPostHook{modifyIdentity: true},
				},
				expectSchemaID: updatedSchema.ID,
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				conf, reg := internal.NewRegistryDefault(t)

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))
				i.TraitsSchemaID = ""
				i.Traits = identity.Traits(`{}`)
				viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/login.schema.json")
				viper.Set(configuration.ViperKeyIdentityTraitsSchemas, []configuration.SchemaConfig{updatedSchema})
				viper.Set(configuration.ViperKeyURLsSelfPublic, "http://example.com")
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.TODO(), &i))

				e := login.NewHookExecutor(reg, conf)
				err := e.PostLoginHook(nil, &http.Request{}, tc.hooks, nil, &i)
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

	t.Run("method=PreLoginHook", func(t *testing.T) {
		for k, tc := range []struct {
			expectErr error
			reg       *loginExecutorDependenciesMock
		}{
			{
				reg:       &loginExecutorDependenciesMock{preErr: []error{nil, nil, errors.New("err")}},
				expectErr: errors.New("err"),
			},
			{reg: &loginExecutorDependenciesMock{preErr: []error{nil}}},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				conf, _ := internal.NewRegistryDefault(t)
				e := login.NewHookExecutor(tc.reg, conf)
				if tc.expectErr == nil {
					require.NoError(t, e.PreLoginHook(nil, nil, nil))
				} else {
					require.EqualError(t, e.PreLoginHook(nil, nil, nil), tc.expectErr.Error())
				}
			})
		}
	})
}
