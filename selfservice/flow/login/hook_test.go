package login_test

import (
	"context"
	"encoding/json"
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

func (m *mockPostHook) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, a *login.Request, s *session.Session) error {
	if m.modifyIdentity {
		i := s.Identity
		i.TraitsSchemaURL = "file://./stub/updated.schema.json"
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

func (m *loginExecutorDependenciesMock) IdentityPool() identity.Pool {
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
			hooks           []login.PostHookExecutor
			expectSchemaURL string
			expectErr       error
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
				expectSchemaURL: "file://./stub/updated.schema.json",
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				conf, reg := internal.NewMemoryRegistry(t)

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))
				i.TraitsSchemaURL = ""
				i.Traits = json.RawMessage(`{}`)
				viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/login.schema.json")
				require.NoError(t, reg.IdentityPool().Create(context.TODO(), &i))

				e := login.NewHookExecutor(reg, conf)
				err = e.PostLoginHook(nil, &http.Request{}, tc.hooks, nil, &i)
				if tc.expectErr != nil {
					require.EqualError(t, err, tc.expectErr.Error())
					return
				}

				require.NoError(t, err)
				if tc.expectSchemaURL != "" {
					got, err := reg.IdentityPool().Get(context.TODO(), i.ID)
					require.NoError(t, err)
					assert.EqualValues(t, tc.expectSchemaURL, got.TraitsSchemaURL)
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
				conf, _ := internal.NewMemoryRegistry(t)
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
