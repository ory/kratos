package settings

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/x"
)

type (
	RequestPersister interface {
		CreateSettingsRequest(context.Context, *Request) error
		GetSettingsRequest(ctx context.Context, id uuid.UUID) (*Request, error)
		UpdateSettingsRequest(context.Context, *Request) error
	}
	RequestPersistenceProvider interface {
		SettingsRequestPersister() RequestPersister
	}
)

func TestRequestPersister(p interface {
	RequestPersister
	identity.PrivilegedPool
}) func(t *testing.T) {
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/identity.schema.json")

	var clearids = func(r *Request) {
		r.ID = uuid.UUID{}
		r.Identity.ID = uuid.UUID{}
		r.IdentityID = uuid.UUID{}
	}

	return func(t *testing.T) {
		t.Run("case=should error when the settings request does not exist", func(t *testing.T) {
			_, err := p.GetSettingsRequest(context.Background(), x.NewUUID())
			require.Error(t, err)
		})

		var newRequest = func(t *testing.T) *Request {
			var r Request
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)
			require.NoError(t, p.CreateIdentity(context.Background(), r.Identity))
			return &r
		}

		t.Run("case=should create a new settings request", func(t *testing.T) {
			r := newRequest(t)
			err := p.CreateSettingsRequest(context.Background(), r)
			require.NoError(t, err, "%#v", err)
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r Request
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateIdentity(context.Background(), r.Identity))
			require.NoError(t, p.CreateSettingsRequest(context.Background(), &r))
		})

		t.Run("case=should create and fetch a settings request", func(t *testing.T) {
			expected := newRequest(t)
			err := p.CreateSettingsRequest(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetSettingsRequest(context.Background(), expected.ID)
			require.NoError(t, err)

			factual, _ := json.Marshal(actual.Methods[StrategyProfile].Config)
			fexpected, _ := json.Marshal(expected.Methods[StrategyProfile].Config)

			require.NotEmpty(t, actual.Methods[StrategyProfile].Config.RequestMethodConfigurator.(*form.HTMLForm).Action)
			assert.EqualValues(t, expected.ID, actual.ID)
			assert.JSONEq(t, string(fexpected), string(factual))
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
			assert.EqualValues(t, expected.Identity.ID, actual.Identity.ID)
			assert.EqualValues(t, expected.Identity.Traits, actual.Identity.Traits)
			assert.EqualValues(t, expected.Identity.TraitsSchemaID, actual.Identity.TraitsSchemaID)
			assert.Empty(t, actual.Identity.Credentials)
		})

		t.Run("case=should fail to create if identity does not exist", func(t *testing.T) {
			var expected Request
			require.NoError(t, faker.FakeData(&expected))
			clearids(&expected)
			err := p.CreateSettingsRequest(context.Background(), &expected)
			require.Error(t, err, "%+s", expected)
		})

		t.Run("case=should create and update a settings request", func(t *testing.T) {
			expected := newRequest(t)
			expected.Methods["oidc"] = &RequestMethod{
				Method: "oidc", Config: &RequestMethodConfig{RequestMethodConfigurator: &form.HTMLForm{Fields: []form.Field{{
					Name: "zab", Type: "bar", Pattern: "baz"}}}}}
			expected.Methods["password"] = &RequestMethod{
				Method: "password", Config: &RequestMethodConfig{RequestMethodConfigurator: &form.HTMLForm{Fields: []form.Field{{
					Name: "foo", Type: "bar", Pattern: "baz"}}}}}
			err := p.CreateSettingsRequest(context.Background(), expected)
			require.NoError(t, err)

			expected.Methods[StrategyProfile].Config.RequestMethodConfigurator.(*form.HTMLForm).Action = "/new-action"
			expected.Methods["password"].Config.RequestMethodConfigurator.(*form.HTMLForm).Fields = []form.Field{{
				Name: "zab", Type: "zab", Pattern: "zab"}}
			expected.RequestURL = "/new-request-url"
			require.NoError(t, p.UpdateSettingsRequest(context.Background(), expected))

			actual, err := p.GetSettingsRequest(context.Background(), expected.ID)
			require.NoError(t, err)

			assert.Equal(t, "/new-action", actual.Methods[StrategyProfile].Config.RequestMethodConfigurator.(*form.HTMLForm).Action)
			assert.Equal(t, "/new-request-url", actual.RequestURL)
			assert.EqualValues(t, []form.Field{{Name: "zab", Type: "zab", Pattern: "zab"}}, actual.
				Methods["password"].Config.RequestMethodConfigurator.(*form.HTMLForm).Fields)
			assert.EqualValues(t, []form.Field{{Name: "zab", Type: "bar", Pattern: "baz"}}, actual.
				Methods["oidc"].Config.RequestMethodConfigurator.(*form.HTMLForm).Fields)
		})
	}
}
