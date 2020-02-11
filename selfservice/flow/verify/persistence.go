package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

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

		TrackAddresses(ctx context.Context, addresses []Address) error

		// FindAddressByCode returns a matching address or sql.ErrNoRows if no address could be found.
		FindAddressByCode(ctx context.Context, code string) (*Address, error)

		// FindAddressByValue returns a matching address or sql.ErrNoRows if no address could be found.
		FindAddressByValue(ctx context.Context, via Via, address string) (*Address, error)

		// VerifyAddress verifies an address by the given id.
		VerifyAddress(ctx context.Context, id uuid.UUID) error

		// VerifyAddress verifies an address by the given id.
		RefreshAddress(ctx context.Context, id uuid.UUID) (*Address, error)
	}
)

func TestPersister(p interface {
	Persister
	identity.Pool
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

			t.Run("case=should create and update a profile request", func(t *testing.T) {
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

		t.Run("suite=address", func(t *testing.T) {
			t.Run("case=not found", func(t *testing.T) {
				_, err := p.FindAddressByCode(context.Background(), "does-not-exist")
				require.Equal(t, sqlcon.ErrNoRows, errorsx.Cause(err))

				_, err = p.FindAddressByValue(context.Background(), ViaEmail, "does-not-exist")
				require.Equal(t, sqlcon.ErrNoRows, errorsx.Cause(err))
			})

			t.Run("case=create and find", func(t *testing.T) {
				addresses := make([]Address, 15)
				for k := range addresses {
					var i identity.Identity
					require.NoError(t, faker.FakeData(&i))
					require.NoError(t, p.CreateIdentity(context.Background(), &i))

					address, err := NewEmailAddress("verify.TestPersister.CreateIdentity"+strconv.Itoa(k)+"@ory.sh", i.ID, time.Minute)
					require.NoError(t, err)
					address.ID = uuid.Nil

					addresses[k] = *address
				}

				require.NoError(t, p.TrackAddresses(context.Background(), addresses))

				for _, address := range addresses {
					require.NotEmpty(t, address.ID)
				}

				compare := func(t *testing.T, expected, actual Address) {
					actual.CreatedAt = actual.CreatedAt.UTC()
					actual.UpdatedAt = actual.UpdatedAt.UTC()
					expected.CreatedAt = expected.CreatedAt.UTC()
					expected.UpdatedAt = expected.UpdatedAt.UTC()
					assert.EqualValues(t, expected, actual)

				}

				for k, expected := range addresses {
					t.Run("method=FindAddressByCode", func(t *testing.T) {
						t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
							actual, err := p.FindAddressByCode(context.Background(), expected.Code)
							require.NoError(t, err)
							compare(t, expected, *actual)
						})
					})

					t.Run("method=FindAddressByValue", func(t *testing.T) {
						t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
							actual, err := p.FindAddressByValue(context.Background(), expected.Via, expected.Value)
							require.NoError(t, err)
							compare(t, expected, *actual)
						})
					})
				}
			})

			t.Run("case=create and verify", func(t *testing.T) {
				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))
				require.NoError(t, p.CreateIdentity(context.Background(), &i))

				address, err := NewEmailAddress("verify.TestPersister.VerifyAddress@ory.sh", i.ID, time.Minute)
				require.NoError(t, err)
				address.ID = x.NewUUID()

				require.NoError(t, p.TrackAddresses(context.Background(), []Address{*address}))
				require.NoError(t, p.VerifyAddress(context.Background(), address.ID))

				actual, err := p.FindAddressByValue(context.Background(), address.Via, address.Value)
				require.NoError(t, err)
				assert.NotEqual(t, address.Code, actual.Code)
				assert.True(t, actual.Verified)
				assert.EqualValues(t, StatusCompleted, actual.Status)
				assert.NotEmpty(t, actual.VerifiedAt)
			})
		})
	}
}
