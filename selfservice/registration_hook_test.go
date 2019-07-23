package selfservice_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
	. "github.com/ory/hive/selfservice"
	"github.com/ory/hive/session"
)

type registrationPostHookMock struct {
	err            error
	modifyIdentity bool
}

func (m *registrationPostHookMock) ExecuteRegistrationPostHook(w http.ResponseWriter, r *http.Request, a *RegistrationRequest, s *session.Session) error {
	if m.modifyIdentity {
		i := s.Identity
		i.TraitsSchemaURL = "updated-schema"
		s.UpdateIdentity(i)
	}
	return m.err
}

type registrationPreHookMock struct {
	err error
}

func (m *registrationPreHookMock) ExecuteRegistrationPreHook(w http.ResponseWriter, r *http.Request, a *RegistrationRequest) error {
	return m.err
}

type registrationExecutorDependenciesMock struct {
	preErr []error
}

func (m *registrationExecutorDependenciesMock) IdentityPool() identity.Pool {
	return nil
}

func (m *registrationExecutorDependenciesMock) AuthHookRegistrationPreExecutors() []HookRegistrationPreExecutor {
	hooks := make([]HookRegistrationPreExecutor, len(m.preErr))
	for k := range hooks {
		hooks[k] = &registrationPreHookMock{m.preErr[k]}
	}
	return hooks
}

func TestRegistrationExecutor(t *testing.T) {
	t.Run("method=PostRegistrationHook", func(t *testing.T) {
		for k, tc := range []struct {
			hooks           []HookRegistrationPostExecutor
			expectSchemaURL string
			expectErr       error
		}{
			{hooks: nil},
			{hooks: []HookRegistrationPostExecutor{}},
			{hooks: []HookRegistrationPostExecutor{&registrationPostHookMock{err: errors.New("err")}}, expectErr: errors.New("err")},
			{hooks: []HookRegistrationPostExecutor{
				new(registrationPostHookMock),
				&registrationPostHookMock{err: errors.New("err")}}, expectErr: errors.New("err"),
			},
			{
				hooks: []HookRegistrationPostExecutor{
					new(registrationPostHookMock),
					&registrationPostHookMock{modifyIdentity: true},
				},
				expectSchemaURL: "updated-schema",
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				conf, reg := internal.NewMemoryRegistry(t)

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				e := NewRegistrationExecutor(reg, conf)
				err := e.PostRegistrationHook(nil, &http.Request{}, tc.hooks, nil, &i)
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

	t.Run("method=PreRegistrationHook", func(t *testing.T) {
		for k, tc := range []struct {
			expectErr error
			reg       *registrationExecutorDependenciesMock
		}{
			{
				reg:       &registrationExecutorDependenciesMock{preErr: []error{nil, nil, errors.New("err")}},
				expectErr: errors.New("err"),
			},
			{
				reg: &registrationExecutorDependenciesMock{preErr: []error{nil}},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				conf, _ := internal.NewMemoryRegistry(t)
				e := NewRegistrationExecutor(tc.reg, conf)
				if tc.expectErr == nil {
					require.NoError(t, e.PreRegistrationHook(nil, nil, nil))
				} else {
					require.EqualError(t, e.PreRegistrationHook(nil, nil, nil), tc.expectErr.Error())
				}
			})
		}
	})
}
