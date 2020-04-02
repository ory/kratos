package verify

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/viper"
	"github.com/ory/x/errorsx"
	"github.com/ory/x/sqlcon"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/x"
)

type (
	PersistenceProvider interface {
		VerificationPersister() Persister
	}
	Persister interface {
		CreateVerifyRequest(context.Context, *Request) error
		GetVerifyRequest(ctx context.Context, id uuid.UUID) (*Request, error)
		UpdateVerifyRequest(context.Context, *Request) error
	}
)

func TestPersister(p interface {
	Persister
	identity.PrivilegedPool
}) func(t *testing.T) {
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/identity.schema.json")
	return func(t *testing.T) {
		t.Run("suite=request", func(t *testing.T) {
			t.Run("case=should error when the verify does not exist", func(t *testing.T) {
				_, err := p.GetVerifyRequest(context.Background(), x.NewUUID())
				require.Equal(t, errorsx.Cause(err), sqlcon.ErrNoRows)
			})

			var clearids = func(r *Request) {
				r.ID = uuid.UUID{}
			}

			var newRequest = func(t *testing.T) *Request {
				var r Request
				require.NoError(t, faker.FakeData(&r))
				r.Via = identity.VerifiableAddressTypeEmail
				clearids(&r)
				return &r
			}

			t.Run("case=should create and fetch verify request", func(t *testing.T) {
				expected := newRequest(t)
				expected.Form = form.NewHTMLForm("some/action")
				err := p.CreateVerifyRequest(context.Background(), expected)
				require.NoError(t, err, "%#v", err)
				actual, err := p.GetVerifyRequest(context.Background(), expected.ID)
				require.NoError(t, err)

				factual, err := json.Marshal(actual.Form)
				require.NoError(t, err)
				fexpected, err := json.Marshal(expected.Form)
				require.NoError(t, err)

				assert.NotEmpty(t, actual.Form.Action)
				assert.EqualValues(t, expected.ID, actual.ID)
				assert.JSONEq(t, string(fexpected), string(factual))
				x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
				x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
				assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
				assert.EqualValues(t, expected.Via, actual.Via)
				assert.EqualValues(t, expected.CSRFToken, actual.CSRFToken)
			})

			t.Run("case=should create and update a verify request", func(t *testing.T) {
				expected := newRequest(t)
				expected.Form = form.NewHTMLForm("some/action")
				err := p.CreateVerifyRequest(context.Background(), expected)
				require.NoError(t, err)

				expected.Form.Action = "/new-action"
				expected.RequestURL = "/new-request-url"
				require.NoError(t, p.UpdateVerifyRequest(context.Background(), expected))

				actual, err := p.GetVerifyRequest(context.Background(), expected.ID)
				require.NoError(t, err)

				assert.Equal(t, "/new-action", actual.Form.Action)
				assert.Equal(t, "/new-request-url", actual.RequestURL)
			})

		})
	}
}
