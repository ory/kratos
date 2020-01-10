package registration_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/viper"
	"net/http"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/flow/registration"
	"github.com/ory/kratos/session"
)

type registrationPostHookMock struct {
	err            error
	modifyIdentity bool
}

func (m *registrationPostHookMock) ExecuteRegistrationPostHook(w http.ResponseWriter, r *http.Request, a *registration.Request, s *session.Session) error {
	if m.modifyIdentity {
		i := s.Identity
		i.Traits = identity.Traits(`{"foo":"bar"}"`)
		s.UpdateIdentity(i)
	}
	return m.err
}

type registrationPreHookMock struct {
	err error
}

func (m *registrationPreHookMock) ExecuteRegistrationPreHook(w http.ResponseWriter, r *http.Request, a *registration.Request) error {
	return m.err
}

type registrationExecutorDependenciesMock struct {
	preErr []error
}

func (m *registrationExecutorDependenciesMock) PostRegistrationHooks(credentialsType identity.CredentialsType) []registration.PostHookExecutor {
	return nil
}

func (m *registrationExecutorDependenciesMock) IdentityPool() identity.Pool {
	return nil
}

func (m *registrationExecutorDependenciesMock) IdentityValidator() *identity.Validator {
	return nil
}

func (m *registrationExecutorDependenciesMock) Logger() logrus.FieldLogger {
	return logrus.New()
}

func (m *registrationExecutorDependenciesMock) PreRegistrationHooks() []registration.PreHookExecutor {
	hooks := make([]registration.PreHookExecutor, len(m.preErr))
	for k := range hooks {
		hooks[k] = &registrationPreHookMock{m.preErr[k]}
	}
	return hooks
}

func TestRegistrationExecutor(t *testing.T) {
	t.Run("method=PostRegistrationHook", func(t *testing.T) {
		for k, tc := range []struct {
			hooks        []registration.PostHookExecutor
			expectTraits string
			expectErr    error
		}{
			{hooks: nil},
			{hooks: []registration.PostHookExecutor{}},
			{hooks: []registration.PostHookExecutor{&registrationPostHookMock{err: errors.New("err")}}, expectErr: errors.New("err")},
			{hooks: []registration.PostHookExecutor{
				new(registrationPostHookMock),
				&registrationPostHookMock{err: errors.New("err")}}, expectErr: errors.New("err"),
			},
			{
				hooks: []registration.PostHookExecutor{
					new(registrationPostHookMock),
					&registrationPostHookMock{modifyIdentity: true},
				},
				expectTraits: `{"foo":"bar"}`,
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				conf, reg := internal.NewRegistryDefault(t)
				_, _ = reg.SchemaPersister().RegisterDefaultSchema("file://./stub/registration.schema.json")
				viper.Set(configuration.ViperKeyURLsSelfPublic, "http://mock-server.com")

				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))
				i.TraitsSchemaID = uuid.Nil
				i.Traits = identity.Traits("{}")

				e := registration.NewHookExecutor(reg, conf)
				err := e.PostRegistrationHook(nil, &http.Request{}, tc.hooks, nil, &i)
				if tc.expectErr != nil {
					require.EqualError(t, err, tc.expectErr.Error())
					return
				}

				require.NoError(t, err)
				if tc.expectTraits != "" {
					got, err := reg.IdentityPool().GetIdentity(context.TODO(), i.ID)
					require.NoError(t, err)
					assert.EqualValues(t, tc.expectTraits, string(got.Traits))
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
				conf, _ := internal.NewRegistryDefault(t)
				e := registration.NewHookExecutor(tc.reg, conf)
				if tc.expectErr == nil {
					require.NoError(t, e.PreRegistrationHook(nil, nil, nil))
				} else {
					require.EqualError(t, e.PreRegistrationHook(nil, nil, nil), tc.expectErr.Error())
				}
			})
		}
	})
}
