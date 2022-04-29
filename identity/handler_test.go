package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ory/kratos/hash"
	"github.com/ory/x/snapshotx"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
)

func TestHandler(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	// Start kratos server
	publicTS, adminTS := testhelpers.NewKratosServerWithCSRF(t, reg)

	mockServerURL := urlx.ParseOrPanic(publicTS.URL)
	defaultSchemaExternalURL := (&schema.Schema{ID: "default"}).SchemaURL(mockServerURL).String()

	conf.MustSet(config.ViperKeyAdminBaseURL, adminTS.URL)
	testhelpers.SetIdentitySchemas(t, conf, map[string]string{
		"default":  "file://./stub/identity.schema.json",
		"customer": "file://./stub/handler/customer.schema.json",
		"employee": "file://./stub/handler/employee.schema.json",
	})

	conf.MustSet(config.ViperKeyPublicBaseURL, mockServerURL.String())

	var get = func(t *testing.T, base *httptest.Server, href string, expectCode int) gjson.Result {
		res, err := base.Client().Get(base.URL + href)
		require.NoError(t, err)
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		require.EqualValues(t, expectCode, res.StatusCode, "%s", body)
		return gjson.ParseBytes(body)
	}

	var remove = func(t *testing.T, base *httptest.Server, href string, expectCode int) {
		req, err := http.NewRequest("DELETE", base.URL+href, nil)
		require.NoError(t, err)

		res, err := base.Client().Do(req)
		require.NoError(t, err)

		require.EqualValues(t, expectCode, res.StatusCode)
	}

	var send = func(t *testing.T, base *httptest.Server, method, href string, expectCode int, send interface{}) gjson.Result {
		var b bytes.Buffer
		if send != nil {
			require.NoError(t, json.NewEncoder(&b).Encode(send))
		}
		req, err := http.NewRequest(method, base.URL+href, &b)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		res, err := base.Client().Do(req)
		require.NoError(t, err)
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		require.EqualValues(t, expectCode, res.StatusCode, "%s", body)
		return gjson.ParseBytes(body)
	}

	t.Run("case=should return an empty list", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				parsed := get(t, ts, "/identities", http.StatusOK)
				require.True(t, parsed.IsArray(), "%s", parsed.Raw)
				assert.Len(t, parsed.Array(), 0)
			})
		}
	})

	t.Run("case=should return 404 on a non-existing resource", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				_ = get(t, ts, "/identities/does-not-exist", http.StatusNotFound)
			})
		}
	})

	t.Run("case=should fail to create an identity because schema id does not exist", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var i identity.AdminCreateIdentityBody
				i.SchemaID = "does-not-exist"
				res := send(t, ts, "POST", "/identities", http.StatusBadRequest, &i)
				assert.Contains(t, res.Get("error.reason").String(), "does-not-exist", "%s", res)

			})
		}
	})

	t.Run("case=should fail to create an entity because schema is not validating", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var i identity.AdminCreateIdentityBody
				i.Traits = []byte(`{"bar":123}`)
				res := send(t, ts, "POST", "/identities", http.StatusBadRequest, &i)
				assert.Contains(t, res.Get("error.reason").String(), "I[#/traits/bar] S[#/properties/traits/properties/bar/type] expected string, but got number")
			})
		}
	})

	t.Run("case=should fail to create an entity with schema_url set", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				res := send(t, ts, "POST", "/identities", http.StatusBadRequest, json.RawMessage(`{"schema_url":"12345","traits":{}}`))
				assert.Contains(t, res.Get("error.message").String(), "schema_url")
			})
		}
	})

	t.Run("case=should create an identity without an ID", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var i identity.AdminCreateIdentityBody
				i.Traits = []byte(`{"bar":"baz"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &i)
				assert.NotEmpty(t, res.Get("id").String(), "%s", res.Raw)
				assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
				assert.Empty(t, res.Get("credentials").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=should create an identity with metadata", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var i identity.AdminCreateIdentityBody
				i.Traits = []byte(`{"bar":"baz"}`)
				i.MetadataPublic = []byte(`{"public":"baz"}`)
				i.MetadataAdmin = []byte(`{"admin":"baz"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &i)
				assert.EqualValues(t, "baz", res.Get("metadata_admin.admin").String(), "%s", res.Raw)
				assert.EqualValues(t, "baz", res.Get("metadata_public.public").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=should be able to import users", func(t *testing.T) {
		ignoreDefault := []string{"id", "schema_url", "state_changed_at", "created_at", "updated_at"}
		t.Run("without any credentials", func(t *testing.T) {
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.AdminCreateIdentityBody{Traits: []byte(`{"email": "import-1@ory.sh"}`)})
			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			snapshotx.SnapshotTExceptMatchingKeys(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), ignoreDefault)
		})

		t.Run("with cleartext password and oidc credentials", func(t *testing.T) {
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.AdminCreateIdentityBody{Traits: []byte(`{"email": "import-2@ory.sh"}`),
				Credentials: &identity.AdminIdentityImportCredentials{
					Password: &identity.AdminIdentityImportCredentialsPassword{
						Config: identity.AdminIdentityImportCredentialsPasswordConfig{
							Password: "123456",
						},
					},
					OIDC: &identity.AdminIdentityImportCredentialsOIDC{
						Config: identity.AdminIdentityImportCredentialsOIDCConfig{
							Providers: []identity.AdminCreateIdentityImportCredentialsOidcProvider{
								{Subject: "import-2", Provider: "google"},
								{Subject: "import-2", Provider: "github"},
							},
						},
					},
				},
			})

			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			snapshotx.SnapshotTExceptMatchingKeys(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), append(ignoreDefault, "hashed_password"))

			require.NoError(t, hash.Compare(ctx, []byte("123456"), []byte(gjson.GetBytes(actual.Credentials[identity.CredentialsTypePassword].Config, "hashed_password").String())))
		})

		t.Run("with pkbdf2 password", func(t *testing.T) {
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.AdminCreateIdentityBody{Traits: []byte(`{"email": "import-3@ory.sh"}`),
				Credentials: &identity.AdminIdentityImportCredentials{Password: &identity.AdminIdentityImportCredentialsPassword{
					Config: identity.AdminIdentityImportCredentialsPasswordConfig{HashedPassword: "$pbkdf2-sha256$i=1000,l=128$e8/arsEf4cvQihdNgqj0Nw$5xQQKNTyeTHx2Ld5/JDE7A"}}}})
			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			snapshotx.SnapshotTExceptMatchingKeys(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), append(ignoreDefault, "hashed_password"))

			require.NoError(t, hash.Compare(ctx, []byte("123456"), []byte(gjson.GetBytes(actual.Credentials[identity.CredentialsTypePassword].Config, "hashed_password").String())))
		})

		t.Run("with bcrypt2 password", func(t *testing.T) {
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.AdminCreateIdentityBody{Traits: []byte(`{"email": "import-4@ory.sh"}`),
				Credentials: &identity.AdminIdentityImportCredentials{Password: &identity.AdminIdentityImportCredentialsPassword{
					Config: identity.AdminIdentityImportCredentialsPasswordConfig{HashedPassword: "$2a$10$ZsCsoVQ3xfBG/K2z2XpBf.tm90GZmtOqtqWcB5.pYd5Eq8y7RlDyq"}}}})
			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			snapshotx.SnapshotTExceptMatchingKeys(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), append(ignoreDefault, "hashed_password"))

			require.NoError(t, hash.Compare(ctx, []byte("123456"), []byte(gjson.GetBytes(actual.Credentials[identity.CredentialsTypePassword].Config, "hashed_password").String())))
		})

		t.Run("with argon2i password", func(t *testing.T) {
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.AdminCreateIdentityBody{Traits: []byte(`{"email": "import-5@ory.sh"}`),
				Credentials: &identity.AdminIdentityImportCredentials{Password: &identity.AdminIdentityImportCredentialsPassword{
					Config: identity.AdminIdentityImportCredentialsPasswordConfig{HashedPassword: "$argon2i$v=19$m=65536,t=3,p=4$STVE4CQ9qQ1dK/j224VMbA$o8b+k5wdHgBqf7ES+aWG2K7Y9diQ6ahEhbW8zcstXGo"}}}})
			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			snapshotx.SnapshotTExceptMatchingKeys(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), append(ignoreDefault, "hashed_password"))

			require.NoError(t, hash.Compare(ctx, []byte("123456"), []byte(gjson.GetBytes(actual.Credentials[identity.CredentialsTypePassword].Config, "hashed_password").String())))
		})

		t.Run("with argon2id password", func(t *testing.T) {
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.AdminCreateIdentityBody{Traits: []byte(`{"email": "import-6@ory.sh"}`),
				Credentials: &identity.AdminIdentityImportCredentials{Password: &identity.AdminIdentityImportCredentialsPassword{
					Config: identity.AdminIdentityImportCredentialsPasswordConfig{HashedPassword: "$argon2id$v=19$m=16,t=2,p=1$bVI1aE1SaTV6SGQ3bzdXdw$fnjCcZYmEPOUOjYXsT92Cg"}}}})
			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			snapshotx.SnapshotTExceptMatchingKeys(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), append(ignoreDefault, "hashed_password"))

			require.NoError(t, hash.Compare(ctx, []byte("123456"), []byte(gjson.GetBytes(actual.Credentials[identity.CredentialsTypePassword].Config, "hashed_password").String())))
		})
	})

	t.Run("case=unable to set ID itself", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				res := send(t, ts, "POST", "/identities", http.StatusBadRequest, json.RawMessage(`{"id":"12345","traits":{}}`))
				assert.Contains(t, res.Raw, "id")
			})
		}
	})

	t.Run("suite=create and update", func(t *testing.T) {
		var i identity.Identity
		var createOidcIdentity = func(t *testing.T, identifier, accessToken, refreshToken, idToken string, encrypt bool) string {
			transform := func(token string) string {
				if !encrypt {
					return token
				}
				c, err := reg.Cipher().Encrypt(context.Background(), []byte(token))
				require.NoError(t, err)
				return c
			}

			iId := x.NewUUID()
			toJson := func(c identity.CredentialsOIDC) []byte {
				out, err := json.Marshal(&c)
				require.NoError(t, err)
				return out
			}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &identity.Identity{
				ID:     iId,
				Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, identifier)),
				Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypeOIDC: {
						Type:        identity.CredentialsTypeOIDC,
						Identifiers: []string{"bar:" + identifier},
						Config: toJson(identity.CredentialsOIDC{Providers: []identity.CredentialsOIDCProvider{
							{
								Subject:             "foo",
								Provider:            "bar",
								InitialAccessToken:  transform(accessToken + "0"),
								InitialRefreshToken: transform(refreshToken + "0"),
								InitialIDToken:      transform(idToken + "0"),
							},
							{
								Subject:             "baz",
								Provider:            "zab",
								InitialAccessToken:  transform(accessToken + "1"),
								InitialRefreshToken: transform(refreshToken + "1"),
								InitialIDToken:      transform(idToken + "1"),
							},
						}}),
					},
					identity.CredentialsTypePassword: {
						Type:        identity.CredentialsTypePassword,
						Identifiers: []string{identifier},
					},
				},
				VerifiableAddresses: []identity.VerifiableAddress{
					{
						ID:         x.NewUUID(),
						Value:      identifier,
						Verified:   false,
						CreatedAt:  time.Now(),
						IdentityID: iId,
					},
				},
			}))
			return iId.String()
		}
		t.Run("case=should create an identity with an ID which is ignored", func(t *testing.T) {
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					res := send(t, ts, "POST", "/identities", http.StatusCreated, json.RawMessage(`{"traits": {"bar":"baz"}}`))
					stateChangedAt := sqlxx.NullTime(res.Get("state_changed_at").Time())

					i.Traits = []byte(res.Get("traits").Raw)
					i.ID = x.ParseUUID(res.Get("id").String())
					i.StateChangedAt = &stateChangedAt
					assert.NotEmpty(t, res.Get("id").String())

					assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
					assert.Empty(t, res.Get("credentials").String(), "%s", res.Raw)
					assert.EqualValues(t, defaultSchemaExternalURL, res.Get("schema_url").String(), "%s", res.Raw)
					assert.EqualValues(t, config.DefaultIdentityTraitsSchemaID, res.Get("schema_id").String(), "%s", res.Raw)
					assert.EqualValues(t, identity.StateActive, res.Get("state").String(), "%s", res.Raw)
				})
			}
		})

		t.Run("case=should be able to get the identity", func(t *testing.T) {
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					res := get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
					assert.EqualValues(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)
					assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
					assert.EqualValues(t, defaultSchemaExternalURL, res.Get("schema_url").String(), "%s", res.Raw)
					assert.EqualValues(t, config.DefaultIdentityTraitsSchemaID, res.Get("schema_id").String(), "%s", res.Raw)
					assert.EqualValues(t, identity.StateActive, res.Get("state").String(), "%s", res.Raw)
					assert.Empty(t, res.Get("credentials").String(), "%s", res.Raw)
				})
			}
		})

		t.Run("case=should get oidc credential", func(t *testing.T) {
			id := createOidcIdentity(t, "foo.oidc@bar.com", "access_token", "refresh_token", "id_token", true)
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					res := get(t, ts, "/identities/"+id, http.StatusOK)
					assert.False(t, res.Get("credentials.oidc.config").Exists(), "credentials config should be omitted: %s", res.Raw)
					assert.False(t, res.Get("credentials.password.config").Exists(), "credentials config should be omitted: %s", res.Raw)

					res = get(t, ts, "/identities/"+id+"?include_credential=oidc", http.StatusOK)
					assert.True(t, res.Get("credentials").Exists(), "credentials should be included: %s", res.Raw)
					assert.True(t, res.Get("credentials.password").Exists(), "password meta should be included: %s", res.Raw)
					assert.False(t, res.Get("credentials.password.false").Exists(), "password credentials should not be included: %s", res.Raw)
					assert.True(t, res.Get("credentials.oidc.config").Exists(), "oidc credentials should be included: %s", res.Raw)

					assert.EqualValues(t, "foo", res.Get("credentials.oidc.config.providers.0.subject").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "bar", res.Get("credentials.oidc.config.providers.0.provider").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "access_token0", res.Get("credentials.oidc.config.providers.0.initial_access_token").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "refresh_token0", res.Get("credentials.oidc.config.providers.0.initial_refresh_token").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "id_token0", res.Get("credentials.oidc.config.providers.0.initial_id_token").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "baz", res.Get("credentials.oidc.config.providers.1.subject").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "zab", res.Get("credentials.oidc.config.providers.1.provider").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "access_token1", res.Get("credentials.oidc.config.providers.1.initial_access_token").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "refresh_token1", res.Get("credentials.oidc.config.providers.1.initial_refresh_token").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "id_token1", res.Get("credentials.oidc.config.providers.1.initial_id_token").String(), "credentials should be included: %s", res.Raw)
				})
			}
		})

		t.Run("case=should pass if no oidc credentials are set", func(t *testing.T) {
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					res := send(t, ts, "POST", "/identities", http.StatusCreated, json.RawMessage(`{"traits": {"bar":"baz"}}`))

					res = get(t, ts, "/identities/"+res.Get("id").String(), http.StatusOK)
					assert.False(t, res.Get("credentials.oidc.config").Exists(), "credentials config should be omitted: %s", res.Raw)
					assert.False(t, res.Get("credentials.password.config").Exists(), "credentials config should be omitted: %s", res.Raw)

					res = get(t, ts, "/identities/"+res.Get("id").String()+"?include_credential=oidc", http.StatusOK)
					assert.False(t, res.Get("credentials.password").Exists(), "password credentials should not be included: %s", res.Raw)
					assert.False(t, res.Get("credentials.oidc").Exists(), "oidc credentials should be included: %s", res.Raw)
				})
			}
		})

		t.Run("case=should fail to get oidc credential", func(t *testing.T) {
			id := createOidcIdentity(t, "foo-failed.oidc@bar.com", "foo_token", "bar_token", "id_token", false)
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					t.Logf("no oidc token")
					res := get(t, ts, "/identities/"+i.ID.String()+"?include_credential=oidc", http.StatusOK)
					assert.NotContains(t, res.Raw, "identifier_credentials", res.Raw)

					t.Logf("get oidc token")
					res = get(t, ts, "/identities/"+id+"?include_credential=oidc", http.StatusInternalServerError)
					assert.Contains(t, res.Raw, "Internal Server Error", res.Raw)
				})
			}

			e, _ := reg.Cipher().Encrypt(context.Background(), []byte("foo_token"))
			id = createOidcIdentity(t, "foo-failed-2.oidc@bar.com", e, "bar_token", "id_token", false)
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					t.Logf("no oidc token")
					res := get(t, ts, "/identities/"+i.ID.String()+"?include_credential=oidc", http.StatusOK)
					assert.NotContains(t, res.Raw, "identifier_credentials", res.Raw)

					t.Logf("get oidc token")
					res = get(t, ts, "/identities/"+id+"?include_credential=oidc", http.StatusInternalServerError)
					assert.Contains(t, res.Raw, "Internal Server Error", res.Raw)
				})
			}
		})

		t.Run("case=should update an identity and persist the changes", func(t *testing.T) {
			i := &identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, x.NewUUID().String()))}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					ur := identity.AdminUpdateIdentityBody{
						Traits:         []byte(`{"bar":"baz","foo":"baz"}`),
						SchemaID:       i.SchemaID,
						State:          identity.StateInactive,
						MetadataPublic: []byte(`{"public":"metadata"}`),
						MetadataAdmin:  []byte(`{"admin":"metadata"}`),
					}

					res := send(t, ts, "PUT", "/identities/"+i.ID.String(), http.StatusOK, &ur)
					assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
					assert.EqualValues(t, "baz", res.Get("traits.foo").String(), "%s", res.Raw)
					assert.EqualValues(t, "metadata", res.Get("metadata_admin.admin").String(), "%s", res.Raw)
					assert.EqualValues(t, "metadata", res.Get("metadata_public.public").String(), "%s", res.Raw)
					assert.EqualValues(t, identity.StateInactive, res.Get("state").String(), "%s", res.Raw)
					assert.NotEqualValues(t, i.StateChangedAt, sqlxx.NullTime(res.Get("state_changed_at").Time()), "%s", res.Raw)

					res = get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
					assert.EqualValues(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)
					assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
					assert.EqualValues(t, "metadata", res.Get("metadata_admin.admin").String(), "%s", res.Raw)
					assert.EqualValues(t, "metadata", res.Get("metadata_public.public").String(), "%s", res.Raw)
					assert.EqualValues(t, identity.StateInactive, res.Get("state").String(), "%s", res.Raw)
					assert.NotEqualValues(t, i.StateChangedAt, sqlxx.NullTime(res.Get("state_changed_at").Time()), "%s", res.Raw)
				})
			}
		})

		t.Run("case=should delete a user and no longer be able to retrieve it", func(t *testing.T) {
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					res := send(t, ts, "POST", "/identities", http.StatusCreated, json.RawMessage(`{"traits": {"bar":"baz"}}`))
					remove(t, ts, "/identities/"+res.Get("id").String(), http.StatusNoContent)
					_ = get(t, ts, "/identities/"+res.Get("id").String(), http.StatusNotFound)
				})
			}
		})
	})

	t.Run("case=should return entity with credentials metadata", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				// create identity with credentials
				i := identity.NewIdentity("")
				i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
					Type:   identity.CredentialsTypePassword,
					Config: sqlxx.JSONRawMessage(`{"secret":"pst"}`),
				})
				i.Traits = identity.Traits("{}")

				require.NoError(t, reg.Persister().CreateIdentity(context.Background(), i))
				res := get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
				assert.EqualValues(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)
				assert.True(t, res.Get("credentials").Exists())
				// Should contain changed date
				assert.True(t, res.Get("credentials.password.updated_at").Exists())
				// Should not contain secrets
				assert.False(t, res.Get("credentials.password.config").Exists())
			})
		}
	})

	t.Run("case=should not be able to create an identity with an invalid schema", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var cr identity.AdminCreateIdentityBody
				cr.SchemaID = "unknown"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh"}`)
				res := send(t, ts, "POST", "/identities", http.StatusBadRequest, &cr)
				assert.Contains(t, res.Raw, "unknown")
			})
		}
	})

	t.Run("case=should not be able to create an identity with an invalid state", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var cr identity.AdminCreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh"}`)
				cr.State = "invalid-state"

				res := send(t, ts, "POST", "/identities", http.StatusBadRequest, &cr)
				assert.Contains(t, res.Get("error.reason").String(), `identity state is not valid`, "%s", res.Raw)
			})
		}
	})

	t.Run("case=should create an identity with a different schema", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var cr identity.AdminCreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh"}`)

				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)
				assert.JSONEq(t, string(cr.Traits), res.Get("traits").Raw, "%s", res.Raw)
				assert.EqualValues(t, "employee", res.Get("schema_id").String(), "%s", res.Raw)
				assert.EqualValues(t, identity.StateActive, res.Get("state").String(), "%s", res.Raw)
				assert.EqualValues(t, mockServerURL.String()+"/schemas/ZW1wbG95ZWU", res.Get("schema_url").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=should create an identity with an explicit active state", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var cr identity.AdminCreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh"}`)
				cr.State = identity.StateActive

				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)
				assert.JSONEq(t, string(cr.Traits), res.Get("traits").Raw, "%s", res.Raw)
				assert.EqualValues(t, "employee", res.Get("schema_id").String(), "%s", res.Raw)
				assert.EqualValues(t, identity.StateActive, res.Get("state").String(), "%s", res.Raw)
				assert.EqualValues(t, mockServerURL.String()+"/schemas/ZW1wbG95ZWU", res.Get("schema_url").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=should create an identity with an explicit inactive state", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var cr identity.AdminCreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh"}`)
				cr.State = identity.StateInactive

				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)
				assert.JSONEq(t, string(cr.Traits), res.Get("traits").Raw, "%s", res.Raw)
				assert.EqualValues(t, "employee", res.Get("schema_id").String(), "%s", res.Raw)
				assert.EqualValues(t, identity.StateInactive, res.Get("state").String(), "%s", res.Raw)
				assert.EqualValues(t, mockServerURL.String()+"/schemas/ZW1wbG95ZWU", res.Get("schema_url").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=should create and sync metadata and update privileged traits", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var cr identity.AdminCreateIdentityBody
				cr.SchemaID = "employee"
				originalEmail := x.NewUUID().String() + "@ory.sh"
				cr.Traits = []byte(`{"email":"` + originalEmail + `"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)
				assert.EqualValues(t, originalEmail, res.Get("recovery_addresses.0.value").String(), "%s", res.Raw)
				assert.EqualValues(t, originalEmail, res.Get("verifiable_addresses.0.value").String(), "%s", res.Raw)

				id := res.Get("id").String()
				updatedEmail := x.NewUUID().String() + "@ory.sh"
				res = send(t, ts, "PUT", "/identities/"+id, http.StatusOK, &identity.AdminUpdateIdentityBody{
					Traits: []byte(`{"email":"` + updatedEmail + `", "department": "ory"}`),
				})

				assert.EqualValues(t, "employee", res.Get("schema_id").String(), "%s", res.Raw)
				assert.EqualValues(t, mockServerURL.String()+"/schemas/ZW1wbG95ZWU", res.Get("schema_url").String(), "%s", res.Raw)
				assert.EqualValues(t, updatedEmail, res.Get("traits.email").String(), "%s", res.Raw)
				assert.EqualValues(t, "ory", res.Get("traits.department").String(), "%s", res.Raw)
				assert.EqualValues(t, updatedEmail, res.Get("recovery_addresses.0.value").String(), "%s", res.Raw)
				assert.EqualValues(t, updatedEmail, res.Get("verifiable_addresses.0.value").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=should update the schema id and fail because traits are invalid", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var cr identity.AdminCreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh", "department": "ory"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)

				id := res.Get("id").String()
				res = send(t, ts, "PUT", "/identities/"+id, http.StatusBadRequest, &identity.AdminUpdateIdentityBody{
					SchemaID: "customer",
					Traits:   cr.Traits,
				})
				assert.Contains(t, res.Get("error.reason").String(), `additionalProperties "department" not allowed`, "%s", res.Raw)
			})
		}
	})

	t.Run("case=should fail to update identity if state is invalid", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var cr identity.AdminCreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh", "department": "ory"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)

				id := res.Get("id").String()
				res = send(t, ts, "PUT", "/identities/"+id, http.StatusBadRequest, &identity.AdminUpdateIdentityBody{
					State:  "invalid-state",
					Traits: []byte(`{"email":"` + faker.Email() + `", "department": "ory"}`),
				})
				assert.Contains(t, res.Get("error.reason").String(), `identity state is not valid`, "%s", res.Raw)
			})
		}
	})

	t.Run("case=should update the schema id", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var cr identity.AdminCreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh", "department": "ory"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)

				id := res.Get("id").String()
				res = send(t, ts, "PUT", "/identities/"+id, http.StatusOK, &identity.AdminUpdateIdentityBody{
					SchemaID: "customer",
					Traits:   []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh", "address": "ory street"}`),
				})
				assert.EqualValues(t, "ory street", res.Get("traits.address").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=should be able to update multiple identities", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				for i := 0; i <= 5; i++ {
					var cr identity.AdminCreateIdentityBody
					cr.SchemaID = "employee"
					cr.Traits = []byte(`{"department": "ory"}`)
					res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)

					id := res.Get("id").String()
					res = send(t, ts, "PUT", "/identities/"+id, http.StatusOK, &identity.AdminUpdateIdentityBody{
						SchemaID: "employee",
						Traits:   []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh"}`),
					})

					res = send(t, ts, "PUT", "/identities/"+id, http.StatusOK, &identity.AdminUpdateIdentityBody{
						SchemaID: "employee",
						Traits:   []byte(`{}`),
					})
				}
			})
		}
	})

	t.Run("case=should fail to update identity if input json is empty or json file does not exist", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var cr identity.AdminCreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh", "department": "ory"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)

				id := res.Get("id").String()
				res = send(t, ts, "PUT", "/identities/"+id, http.StatusBadRequest, nil)
				assert.Contains(t, res.Get("error.reason").String(), `Unable to decode HTTP Request Body because its HTTP `+
					`Header "Content-Length" is zero`, "%s", res.Raw)
			})
		}
	})

	t.Run("case=should list all identities", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				res := get(t, ts, "/identities", http.StatusOK)
				assert.Empty(t, res.Get("0.credentials").String(), "%s", res.Raw)
				assert.EqualValues(t, "baz", res.Get(`#(traits.bar=="baz").traits.bar`).String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=should not be able to update an identity that does not exist yet", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				res := send(t, ts, "PUT", "/identities/not-found", http.StatusNotFound, json.RawMessage(`{"traits": {"bar":"baz"}}`))
				assert.Contains(t, res.Get("error.message").String(), "Unable to locate the resource", "%s", res.Raw)
			})
		}
	})

	t.Run("case=should return 404 for non-existing identities", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				remove(t, ts, "/identities/"+x.NewUUID().String(), http.StatusNotFound)
			})
		}
	})
}
