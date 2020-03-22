package profile

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bxcodec/faker"
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
		CreateProfileRequest(context.Context, *Request) error
		GetProfileRequest(ctx context.Context, id uuid.UUID) (*Request, error)
		UpdateProfileRequest(context.Context, *Request) error
	}
	RequestPersistenceProvider interface {
		ProfileRequestPersister() RequestPersister
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
		t.Run("case=should error when the profile request does not exist", func(t *testing.T) {
			_, err := p.GetProfileRequest(context.Background(), x.NewUUID())
			require.Error(t, err)
		})

		var newRequest = func(t *testing.T) *Request {
			var r Request
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)
			require.NoError(t, p.CreateIdentity(context.Background(), r.Identity))
			return &r
		}

		t.Run("case=should create a new profile request", func(t *testing.T) {
			r := newRequest(t)
			err := p.CreateProfileRequest(context.Background(), r)
			require.NoError(t, err, "%#v", err)
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r Request
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateIdentity(context.Background(), r.Identity))
			require.NoError(t, p.CreateProfileRequest(context.Background(), &r))
		})

		t.Run("case=should create and fetch a profile request", func(t *testing.T) {
			expected := newRequest(t)
			err := p.CreateProfileRequest(context.Background(), expected)
			require.NoError(t, err)

			actual, err := p.GetProfileRequest(context.Background(), expected.ID)
			require.NoError(t, err)

			factual, _ := json.Marshal(actual.Methods[StrategyTraitsID].Config)
			fexpected, _ := json.Marshal(expected.Methods[StrategyTraitsID].Config)

			require.NotEmpty(t, actual.Methods[StrategyTraitsID].Config.RequestMethodConfigurator.(*form.HTMLForm).Action)
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
			err := p.CreateProfileRequest(context.Background(), &expected)
			require.Error(t, err)
		})

		t.Run("case=should create and update a profile request", func(t *testing.T) {
			expected := newRequest(t)
			err := p.CreateProfileRequest(context.Background(), expected)
			require.NoError(t, err)

			expected.Methods[StrategyTraitsID].Config.RequestMethodConfigurator.(*form.HTMLForm).Action = "/new-action"
			expected.RequestURL = "/new-request-url"
			require.NoError(t, p.UpdateProfileRequest(context.Background(), expected))

			actual, err := p.GetProfileRequest(context.Background(), expected.ID)
			require.NoError(t, err)

			assert.Equal(t, "/new-action", actual.Methods[StrategyTraitsID].Config.RequestMethodConfigurator.(*form.HTMLForm).Action)
			assert.Equal(t, "/new-request-url", actual.RequestURL)
		})
	}
}
