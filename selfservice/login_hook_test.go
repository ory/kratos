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

type loginPreHookMock struct {
	err error
}

func (m *loginPreHookMock) ExecuteLoginPreHook(w http.ResponseWriter, r *http.Request, a *LoginRequest) error {
	return m.err
}

type loginPostHookMock struct {
	err            error
	modifyIdentity bool
}

func (m *loginPostHookMock) ExecuteLoginPostHook(w http.ResponseWriter, r *http.Request, a *LoginRequest, s *session.Session) error {
	if m.modifyIdentity {
		i := s.Identity
		i.TraitsSchemaURL = "updated-schema"
		s.UpdateIdentity(i)
	}
	return m.err
}

type loginExecutorDependenciesMock struct {
	preErr []error
}

func (m *loginExecutorDependenciesMock) IdentityPool() identity.Pool {
	return nil
}

func (m *loginExecutorDependenciesMock) AuthHookLoginPreExecutors() []HookLoginPreExecutor {
	hooks := make([]HookLoginPreExecutor, len(m.preErr))
	for k := range hooks {
		hooks[k] = &loginPreHookMock{m.preErr[k]}
	}
	return hooks
}

func TestLoginExecutor(t *testing.T) {
	t.Run("method=PostLoginHook", func(t *testing.T) {
		for k, tc := range []struct {
			hooks           []HookLoginPostExecutor
			expectSchemaURL string
			expectErr       error
		}{
			{hooks: nil},
			{hooks: []HookLoginPostExecutor{}},
			{hooks: []HookLoginPostExecutor{&loginPostHookMock{err: errors.New("err")}}, expectErr: errors.New("err")},
			{hooks: []HookLoginPostExecutor{
				new(loginPostHookMock),
				&loginPostHookMock{err: errors.New("err")}}, expectErr: errors.New("err"),
			},
			{
				hooks: []HookLoginPostExecutor{
					new(loginPostHookMock),
					&loginPostHookMock{modifyIdentity: true},
				},
				expectSchemaURL: "updated-schema",
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				conf, reg := internal.NewMemoryRegistry(t)

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))
				_, err := reg.IdentityPool().Create(context.TODO(), &i)
				require.NoError(t, err)

				e := NewLoginExecutor(reg, conf)
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
				e := NewLoginExecutor(tc.reg, conf)
				if tc.expectErr == nil {
					require.NoError(t, e.PreLoginHook(nil, nil, nil))
				} else {
					require.EqualError(t, e.PreLoginHook(nil, nil, nil), tc.expectErr.Error())
				}
			})
		}
	})
}
