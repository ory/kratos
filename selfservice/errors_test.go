package selfservice_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/gojsonschema"
	"github.com/ory/viper"
	"github.com/ory/x/urlx"

	"github.com/ory/hive/driver/configuration"
	"github.com/ory/hive/identity"
	"github.com/ory/hive/internal"
	"github.com/ory/hive/schema"
	. "github.com/ory/hive/selfservice"
	"github.com/ory/hive/selfservice/password"
)

func TestErrorHandler(t *testing.T) {
	conf, reg := internal.NewMemoryRegistry(t)
	eh := NewErrorHandler(reg, conf)

	viper.Set(configuration.ViperKeyURLsError, "http://hive.ory.sh/error")
	viper.Set(configuration.ViperKeyURLsLogin, "http://hive.ory.sh/form")
	viper.Set(configuration.ViperKeyURLsRegistration, "http://hive.ory.sh/form")

	// cem = config error message
	var cem = func(config RequestMethodConfig) string {
		return config.(*password.RequestMethodConfig).Errors[0].Message
	}

	var assertFormFields = func(t *testing.T, config RequestMethodConfig) {
		assert.Equal(t, "bar", config.GetFormFields()["string"].Value)
		assert.Equal(t, int64(123), config.GetFormFields()["number"].Value)
		assert.Equal(t, false, config.GetFormFields()["boolf"].Value)
		assert.Equal(t, true, config.GetFormFields()["boolt"].Value)
		assert.Equal(t, true, config.GetFormFields()["checkbox"].Value)
	}

	for k, tc := range []struct {
		o                 *ErrorHandlerOptions
		err               error
		endURL            string
		assertConfig      func(t *testing.T, config RequestMethodConfig)
		assertSystemError func(t *testing.T, errs []error)
	}{
		{
			err:    errors.WithStack(ErrIDTokenMissing),
			endURL: "http://hive.ory.sh/form?request=0",
			assertConfig: func(t *testing.T, config RequestMethodConfig) {
				assert.EqualValues(t, ErrIDTokenMissing.Reason(), cem(config))
			},
		},
		{
			err:    errors.WithStack(ErrScopeMissing),
			endURL: "http://hive.ory.sh/form?request=1",
			assertConfig: func(t *testing.T, config RequestMethodConfig) {
				assert.EqualValues(t, ErrScopeMissing.Reason(), cem(config))
			},
		},
		{
			err:    errors.WithStack(ErrLoginRequestExpired),
			endURL: "http://hive.ory.sh/form?request=2",
			assertConfig: func(t *testing.T, config RequestMethodConfig) {
				assert.EqualValues(t, ErrLoginRequestExpired.Reason(), cem(config))
			},
		},
		{
			err:    errors.WithStack(ErrRegistrationRequestExpired),
			endURL: "http://hive.ory.sh/form?request=3",
			assertConfig: func(t *testing.T, config RequestMethodConfig) {
				assert.EqualValues(t, ErrRegistrationRequestExpired.Reason(), cem(config))
			},
		},
		{
			err:    errors.New("some error"),
			endURL: "http://hive.ory.sh/error?error=",
			assertSystemError: func(t *testing.T, errs []error) {
				assert.Len(t, errs, 1)
				assert.EqualError(t, errs[0], "some error")
			},
		},
		{
			err:    errors.WithStack(schema.NewInvalidCredentialsError()),
			endURL: "http://hive.ory.sh/form?request=5",
			assertConfig: func(t *testing.T, config RequestMethodConfig) {
				assertFormFields(t, config)
				assert.EqualValues(t, `The provided credentials are invalid. Check for spelling mistakes in your password or username, email address, or phone number.`, cem(config))
			},
		},
		{
			err:    errors.WithStack(schema.NewRequiredError(nil, gojsonschema.NewJsonContext("field_missing", nil))),
			endURL: "http://hive.ory.sh/form?request=6",
			assertConfig: func(t *testing.T, config RequestMethodConfig) {
				assertFormFields(t, config)
				assert.Contains(t, config.GetFormFields()["field_missing"].Error.Message, "field_missing is required")
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			id := strconv.Itoa(k)
			r := &http.Request{
				Header: map[string][]string{},
				PostForm: url.Values{
					"string":   {"bar"},
					"number":   {"123"},
					"boolf":    {"false"},
					"boolt":    {"true"},
					"checkbox": {"false", "true"},
				},
				URL: urlx.ParseOrPanic("http://hive.ory.sh/form"),
			}

			t.Run("case=form", func(t *testing.T) {
				t.Run("case=login", func(t *testing.T) {
					lr := &LoginRequest{
						Request: &Request{
							ID: id,
							Methods: map[identity.CredentialsType]*DefaultRequestMethod{
								password.CredentialsType: {
									Method: password.CredentialsType,
									Config: &password.RequestMethodConfig{
										Fields: map[string]FormField{},
									},
								},
							},
						},
					}

					require.NoError(t, reg.LoginRequestManager().CreateLoginRequest(context.Background(), lr))

					w := httptest.NewRecorder()
					eh.HandleLoginError(w, r, password.CredentialsType, lr, tc.err, tc.o)
					require.Contains(t, w.Header().Get("Location"), tc.endURL, "%+v", w.Header())

					if tc.assertConfig != nil {
						got, err := reg.LoginRequestManager().GetLoginRequest(context.Background(), id)
						require.NoError(t, err)
						tc.assertConfig(t, got.Methods[password.CredentialsType].Config)
					}

					if tc.assertSystemError != nil {
						errs, err := reg.ErrorManager().Read(urlx.ParseOrPanic(w.Header().Get("Location")).Query().Get("error"))
						require.NoError(t, err)
						tc.assertSystemError(t, errs)
					}
				})

				t.Run("case=registration", func(t *testing.T) {
					rr := &RegistrationRequest{
						Request: &Request{
							ID: id,
							Methods: map[identity.CredentialsType]*DefaultRequestMethod{
								password.CredentialsType: {
									Method: password.CredentialsType,
									Config: &password.RequestMethodConfig{
										Fields: map[string]FormField{},
									},
								},
							},
						},
					}

					require.NoError(t, reg.RegistrationRequestManager().CreateRegistrationRequest(context.Background(), rr))

					w := httptest.NewRecorder()
					eh.HandleRegistrationError(w, r, password.CredentialsType, rr, tc.err, tc.o)
					require.Contains(t, w.Header().Get("Location"), tc.endURL, "%+v", w.Header())

					if tc.assertConfig != nil {
						got, err := reg.RegistrationRequestManager().GetRegistrationRequest(context.Background(), id)
						require.NoError(t, err)
						tc.assertConfig(t, got.Methods[password.CredentialsType].Config)
					}

					if tc.assertSystemError != nil {
						errs, err := reg.ErrorManager().Read(urlx.ParseOrPanic(w.Header().Get("Location")).Query().Get("error"))
						require.NoError(t, err)
						tc.assertSystemError(t, errs)
					}
				})
			})

			t.Run("case=json", func(t *testing.T) {
				t.Skip("see https://github.com/ory/hive/issues/61")
			})
		})
	}
}
