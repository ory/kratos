package test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal/testhelpers"

	"github.com/ory/kratos/persistence"
	"github.com/ory/kratos/selfservice/flow/settings"
	"github.com/ory/x/sqlcon"

	"github.com/ory/x/assertx"

	"github.com/ory/kratos/ui/node"

	"github.com/bxcodec/faker/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/x"
)

func TestRequestPersister(ctx context.Context, conf *config.Config, p interface {
	persistence.Persister
}) func(t *testing.T) {
	var clearids = func(r *settings.Flow) {
		r.ID = uuid.UUID{}
		r.Identity.ID = uuid.UUID{}
		r.IdentityID = uuid.UUID{}
	}

	return func(t *testing.T) {
		_, p := testhelpers.NewNetwork(t, p)

		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

		t.Run("case=should error when the settings request does not exist", func(t *testing.T) {
			_, err := p.GetSettingsFlow(ctx, x.NewUUID())
			require.Error(t, err)
		})

		var newFlow = func(t *testing.T) *settings.Flow {
			var r settings.Flow
			require.NoError(t, faker.FakeData(&r))
			clearids(&r)
			require.NoError(t, p.CreateIdentity(ctx, r.Identity))
			return &r
		}

		t.Run("case=should create a new settings request", func(t *testing.T) {
			r := newFlow(t)
			err := p.CreateSettingsFlow(ctx, r)
			require.NoError(t, err, "%#v", err)

			t.Run("can not fetch on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, p)
				_, err := p.GetSettingsFlow(ctx, r.ID)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})

		t.Run("case=should create with set ids", func(t *testing.T) {
			var r settings.Flow
			require.NoError(t, faker.FakeData(&r))
			require.NoError(t, p.CreateIdentity(ctx, r.Identity))
			require.NoError(t, p.CreateSettingsFlow(ctx, &r))

			require.NotEmpty(t, r.Identity)
			t.Run("can not fetch on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, p)
				_, err := p.GetSettingsFlow(ctx, r.ID)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})

		t.Run("case=should create and fetch a settings request", func(t *testing.T) {
			expected := newFlow(t)
			err := p.CreateSettingsFlow(ctx, expected)
			require.NoError(t, err)

			actual, err := p.GetSettingsFlow(ctx, expected.ID)
			require.NoError(t, err)

			factual, _ := json.Marshal(actual.UI)
			fexpected, _ := json.Marshal(expected.UI)

			require.NotEmpty(t, actual.UI.Action)
			assert.EqualValues(t, expected.ID, actual.ID)
			assert.JSONEq(t, string(fexpected), string(factual))
			x.AssertEqualTime(t, expected.IssuedAt, actual.IssuedAt)
			x.AssertEqualTime(t, expected.ExpiresAt, actual.ExpiresAt)
			assert.EqualValues(t, expected.RequestURL, actual.RequestURL)
			assert.EqualValues(t, expected.Identity.ID, actual.Identity.ID)
			assert.EqualValues(t, expected.Identity.Traits, actual.Identity.Traits)
			assert.EqualValues(t, expected.Identity.SchemaID, actual.Identity.SchemaID)
			assert.Empty(t, actual.Identity.Credentials)
		})

		t.Run("case=should fail to create if identity does not exist", func(t *testing.T) {
			var expected settings.Flow
			require.NoError(t, faker.FakeData(&expected))
			clearids(&expected)
			expected.Identity = nil
			expected.IdentityID = uuid.Nil
			err := p.CreateSettingsFlow(ctx, &expected)
			require.Error(t, err, "%+s", expected)
		})

		t.Run("case=should create and update a settings request", func(t *testing.T) {
			expected := newFlow(t)
			expected.UI.Nodes = node.Nodes{}
			expected.UI.Nodes.Append(node.NewInputField("zab", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
				a.Pattern = "baz"
			})))

			expected.UI.Nodes.Append(node.NewInputField("foo", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
				a.Pattern = "baz"
			})))

			err := p.CreateSettingsFlow(ctx, expected)
			require.NoError(t, err)

			expected.UI.Action = "/new-action"
			expected.UI.Nodes.Append(
				node.NewInputField("zab", nil, node.DefaultGroup, "zab", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "zab"
				})))

			expected.RequestURL = "/new-request-url"
			require.NoError(t, p.UpdateSettingsFlow(ctx, expected))

			actual, err := p.GetSettingsFlow(ctx, expected.ID)
			require.NoError(t, err)

			assert.Equal(t, "/new-action", actual.UI.Action)
			assert.Equal(t, "/new-request-url", actual.RequestURL)
			assertx.EqualAsJSON(t, node.Nodes{
				// v0.5: {Name: "zab", Type: "zab", Pattern: "zab"},
				node.NewInputField("zab", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "baz"
				})),
				node.NewInputField("foo", nil, node.DefaultGroup, "bar", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "baz"
				})),
				// v0.5: {Name: "zab", Type: "bar", Pattern: "baz"},
				node.NewInputField("zab", nil, node.DefaultGroup, "zab", node.WithInputAttributes(func(a *node.InputAttributes) {
					a.Pattern = "zab"
				})),
			}, actual.UI.Nodes)
		})

		t.Run("case=network", func(t *testing.T) {
			id := x.NewUUID()
			iid := x.NewUUID()
			nid, p := testhelpers.NewNetwork(t, p)

			require.NoError(t, p.CreateIdentity(ctx, &identity.Identity{ID: iid}))

			t.Run("sets id on creation", func(t *testing.T) {
				expected := &settings.Flow{ID: id, IdentityID: iid, IssuedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}
				require.NoError(t, p.CreateSettingsFlow(ctx, expected))
				assert.EqualValues(t, id, expected.ID)
				assert.EqualValues(t, nid, expected.NID)

				actual, err := p.GetSettingsFlow(ctx, id)
				require.NoError(t, err)
				assert.EqualValues(t, id, actual.ID)
				assert.EqualValues(t, nid, actual.NID)

				_, err = p.GetSettingsFlow(ctx, id)
				require.NoError(t, err)
			})

			t.Run("can not update id", func(t *testing.T) {
				expected, err := p.GetSettingsFlow(ctx, id)
				require.NoError(t, err)

				_, other := testhelpers.NewNetwork(t, p)

				expected.RequestURL = "updated"
				require.ErrorIs(t, other.UpdateSettingsFlow(ctx, expected), sqlcon.ErrNoRows)

				actual, err := p.GetSettingsFlow(ctx, id)
				require.NoError(t, err)
				assert.EqualValues(t, id, actual.ID)
				assert.EqualValues(t, nid, actual.NID)
				require.NotEqual(t, "updated", actual.RequestURL)
			})

			t.Run("can not get on another network", func(t *testing.T) {
				_, p := testhelpers.NewNetwork(t, p)
				_, err := p.GetSettingsFlow(ctx, id)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})

			t.Run("network isolation", func(t *testing.T) {
				nid2, _ := testhelpers.NewNetwork(t, p)
				nid1, p := testhelpers.NewNetwork(t, p)

				iid1, iid2 := x.NewUUID(), x.NewUUID()
				require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at) VALUES (?, ?, 'default', '{}', ?, ?)", iid1, nid1, time.Now(), time.Now()).Exec())
				require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO identities (id, nid, schema_id, traits, created_at, updated_at) VALUES (?, ?, 'default', '{}', ?, ?)", iid2, nid2, time.Now(), time.Now()).Exec())

				sid1, sid2 := x.NewUUID(), x.NewUUID()
				require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO selfservice_settings_flows (id, nid, identity_id, ui, created_at, updated_at, expires_at, request_url) VALUES (?, ?, ?, '{}', ?, ?, ?, '')", sid1, nid1, iid1, time.Now(), time.Now(), time.Now().Add(time.Hour)).Exec())
				require.NoError(t, p.GetConnection(ctx).RawQuery("INSERT INTO selfservice_settings_flows (id, nid, identity_id, ui, created_at, updated_at, expires_at, request_url) VALUES (?, ?, ?, '{}', ?, ?, ?, '')", sid2, nid2, iid2, time.Now(), time.Now(), time.Now().Add(time.Hour)).Exec())

				_, err := p.GetSettingsFlow(ctx, sid1)
				require.NoError(t, err)
				_, err = p.GetSettingsFlow(ctx, sid2)
				require.ErrorIs(t, err, sqlcon.ErrNoRows)
			})
		})
	}
}
