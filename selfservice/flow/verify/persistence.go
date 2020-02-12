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
			createIdentityWithAddresses := func(t *testing.T, expiry time.Duration, email string) identity.VerifiableAddress {
				var i identity.Identity
				require.NoError(t, faker.FakeData(&i))

				address, err := identity.NewVerifiableEmailAddress(email, i.ID, expiry)
				require.NoError(t, err)
				i.Addresses = append(i.Addresses, *address)

				require.NoError(t, p.CreateIdentity(context.Background(), &i))
				return i.Addresses[0]
			}

			t.Run("case=not found", func(t *testing.T) {
				_, err := p.FindAddressByCode(context.Background(), "does-not-exist")
				require.Equal(t, sqlcon.ErrNoRows, errorsx.Cause(err))

				_, err = p.FindAddressByValue(context.Background(), identity.VerifiableAddressTypeEmail, "does-not-exist")
				require.Equal(t, sqlcon.ErrNoRows, errorsx.Cause(err))
			})

			t.Run("case=create and find", func(t *testing.T) {
				addresses := make([]identity.VerifiableAddress, 15)
				for k := range addresses {
					addresses[k] = createIdentityWithAddresses(t, time.Minute, "verify.TestPersister.Create"+strconv.Itoa(k)+"@ory.sh")
					require.NotEmpty(t, addresses[k].ID)
				}

				compare := func(t *testing.T, expected, actual identity.VerifiableAddress) {
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
				address := createIdentityWithAddresses(t, time.Minute, "verify.TestPersister.VerifyAddress@ory.sh")
				require.NoError(t, p.VerifyAddress(context.Background(), address.Code))

				actual, err := p.FindAddressByValue(context.Background(), address.Via, address.Value)
				require.NoError(t, err)
				assert.NotEqual(t, address.Code, actual.Code)
				assert.True(t, actual.Verified)
				assert.EqualValues(t, identity.VerifiableAddressStatusCompleted, actual.Status)
				assert.NotEmpty(t, actual.VerifiedAt)
			})

			t.Run("case=update", func(t *testing.T) {
				address := createIdentityWithAddresses(t, time.Minute, "verify.TestPersister.Update@ory.sh")

				address.Code = "new-code"
				require.NoError(t, p.UpdateVerifiableAddress(context.Background(), &address))

				actual, err := p.FindAddressByValue(context.Background(), address.Via, address.Value)
				require.NoError(t, err)
				assert.Equal(t, "new-code", actual.Code)
			})
		})
	}
}
