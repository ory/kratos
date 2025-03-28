// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/peterhellberg/link"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/schema"
	"github.com/ory/kratos/x"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/randx"
	"github.com/ory/x/snapshotx"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

func TestHandler(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	// Start kratos server
	publicTS, adminTS := testhelpers.NewKratosServerWithCSRF(t, reg)

	mockServerURL := urlx.ParseOrPanic(publicTS.URL)
	defaultSchemaExternalURL := (&schema.Schema{ID: "default"}).SchemaURL(mockServerURL).String()

	conf.MustSet(ctx, config.ViperKeyAdminBaseURL, adminTS.URL)
	testhelpers.SetIdentitySchemas(t, conf, map[string]string{
		"default":         "file://./stub/identity.schema.json",
		"customer":        "file://./stub/handler/customer.schema.json",
		"multiple_emails": "file://./stub/handler/multiple_emails.schema.json",
		"employee":        "file://./stub/handler/employee.schema.json",
	})

	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, mockServerURL.String())

	getFull := func(t *testing.T, base *httptest.Server, href string, expectCode int) (gjson.Result, *http.Response) {
		t.Helper()
		res, err := base.Client().Get(base.URL + href)
		require.NoError(t, err)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		require.EqualValues(t, expectCode, res.StatusCode, "%s", body)
		return gjson.ParseBytes(body), res
	}

	get := func(t *testing.T, base *httptest.Server, href string, expectCode int) gjson.Result {
		t.Helper()
		res, _ := getFull(t, base, href, expectCode)
		return res
	}

	remove := func(t *testing.T, base *httptest.Server, href string, expectCode int) {
		t.Helper()
		req, err := http.NewRequest("DELETE", base.URL+href, nil)
		require.NoError(t, err)

		res, err := base.Client().Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		require.EqualValues(t, expectCode, res.StatusCode, "%s", ioutilx.MustReadAll(res.Body))
	}

	send := func(t *testing.T, base *httptest.Server, method, href string, expectCode int, send interface{}) gjson.Result {
		t.Helper()
		var b bytes.Buffer
		switch raw := send.(type) {
		case json.RawMessage:
			b = *bytes.NewBuffer(raw)
		default:
			if send != nil {
				require.NoError(t, json.NewEncoder(&b).Encode(send))
			}
		}

		req, err := http.NewRequest(method, base.URL+href, &b)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		res, err := base.Client().Do(req)
		require.NoError(t, err)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		require.EqualValues(t, expectCode, res.StatusCode, "%s", body)
		return gjson.ParseBytes(body)
	}

	type patch map[string]interface{}

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
				var i identity.CreateIdentityBody
				i.SchemaID = "does-not-exist"
				res := send(t, ts, "POST", "/identities", http.StatusBadRequest, &i)
				assert.Contains(t, res.Get("error.reason").String(), "does-not-exist", "%s", res)
			})
		}
	})

	t.Run("case=should fail to create an entity because schema is not validating", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var i identity.CreateIdentityBody
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
				var i identity.CreateIdentityBody
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
				var i identity.CreateIdentityBody
				i.Traits = []byte(`{"bar":"baz"}`)
				i.MetadataPublic = []byte(`{"public":"baz"}`)
				i.MetadataAdmin = []byte(`{"admin":"baz"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &i)
				assert.EqualValues(t, "baz", res.Get("metadata_admin.admin").String(), "%s", res.Raw)
				assert.EqualValues(t, "baz", res.Get("metadata_public.public").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=should create an identity with an organization ID", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				orgID := uuid.NullUUID{x.NewUUID(), true}
				i := identity.CreateIdentityBody{
					Traits:         []byte(`{"bar":"baz"}`),
					OrganizationID: orgID,
				}
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &i)
				assert.EqualValues(t, orgID.UUID.String(), res.Get("organization_id").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=should be able to import users", func(t *testing.T) {
		ignoreDefault := []string{"id", "schema_url", "state_changed_at", "created_at", "updated_at"}
		t.Run("without any credentials", func(t *testing.T) {
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.CreateIdentityBody{Traits: []byte(`{"email": "import-1@ory.sh"}`)})
			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			snapshotx.SnapshotT(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), snapshotx.ExceptNestedKeys(ignoreDefault...))
		})

		t.Run("without traits", func(t *testing.T) {
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, json.RawMessage("{}"))
			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			snapshotx.SnapshotT(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), snapshotx.ExceptNestedKeys(ignoreDefault...))
		})

		t.Run("with malformed traits", func(t *testing.T) {
			send(t, adminTS, "POST", "/identities", http.StatusBadRequest, json.RawMessage(`{"traits": not valid JSON}`))
		})

		t.Run("with cleartext password and oidc credentials", func(t *testing.T) {
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.CreateIdentityBody{
				Traits: []byte(`{"email": "import-2@ory.sh"}`),
				Credentials: &identity.IdentityWithCredentials{
					Password: &identity.AdminIdentityImportCredentialsPassword{
						Config: identity.AdminIdentityImportCredentialsPasswordConfig{
							Password: "123456",
						},
					},
					OIDC: &identity.AdminIdentityImportCredentialsOIDC{
						Config: identity.AdminIdentityImportCredentialsOIDCConfig{
							Providers: []identity.AdminCreateIdentityImportCredentialsOIDCProvider{
								{Subject: "import-2", Provider: "google"},
								{Subject: "import-2", Provider: "github"},
							},
						},
					},
					SAML: &identity.AdminIdentityImportCredentialsSAML{
						Config: identity.AdminIdentityImportCredentialsSAMLConfig{
							Providers: []identity.AdminCreateIdentityImportCredentialsSAMLProvider{
								{Subject: "import-saml-2", Provider: "okta"},
								{Subject: "import-saml-2", Provider: "onelogin"},
							},
						},
					},
				},
			})

			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			snapshotx.SnapshotT(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), snapshotx.ExceptNestedKeys(append(ignoreDefault, "hashed_password")...), snapshotx.ExceptPaths("credentials.oidc.identifiers"))

			identifiers := actual.Credentials[identity.CredentialsTypeOIDC].Identifiers
			assert.Len(t, identifiers, 2)
			assert.Contains(t, identifiers, "google:import-2")
			assert.Contains(t, identifiers, "github:import-2")

			identifiers = actual.Credentials[identity.CredentialsTypeSAML].Identifiers
			assert.Len(t, identifiers, 2)
			assert.Contains(t, identifiers, "okta:import-saml-2")
			assert.Contains(t, identifiers, "onelogin:import-saml-2")

			require.NoError(t, hash.Compare(ctx, []byte("123456"), []byte(gjson.GetBytes(actual.Credentials[identity.CredentialsTypePassword].Config, "hashed_password").String())))
		})

		t.Run("with organization oidc and saml credentials", func(t *testing.T) {
			org := "ad6a7dac-4eef-4f09-8e58-c099c14b6c36"
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.CreateIdentityBody{
				Traits: []byte(`{"email": "import-3@ory.sh"}`),
				Credentials: &identity.IdentityWithCredentials{
					OIDC: &identity.AdminIdentityImportCredentialsOIDC{
						Config: identity.AdminIdentityImportCredentialsOIDCConfig{
							Providers: []identity.AdminCreateIdentityImportCredentialsOIDCProvider{
								{Subject: "import-org-3", Provider: "google", Organization: uuid.NullUUID{Valid: true, UUID: uuid.FromStringOrNil(org)}},
								{Subject: "import-org-3", Provider: "github", Organization: uuid.NullUUID{Valid: true, UUID: uuid.FromStringOrNil(org)}},
							},
						},
					},
					SAML: &identity.AdminIdentityImportCredentialsSAML{
						Config: identity.AdminIdentityImportCredentialsSAMLConfig{
							Providers: []identity.AdminCreateIdentityImportCredentialsSAMLProvider{
								{Subject: "import-saml-org-3", Provider: "okta", Organization: uuid.NullUUID{Valid: true, UUID: uuid.FromStringOrNil(org)}},
								{Subject: "import-saml-org-3", Provider: "onelogin", Organization: uuid.NullUUID{Valid: true, UUID: uuid.FromStringOrNil(org)}},
							},
						},
					},
				},
			})

			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			snapshotx.SnapshotT(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), snapshotx.ExceptNestedKeys(append(ignoreDefault, "hashed_password")...), snapshotx.ExceptPaths("credentials.oidc.identifiers"))

			identifiers := actual.Credentials[identity.CredentialsTypeOIDC].Identifiers
			assert.Len(t, identifiers, 2)
			assert.Contains(t, identifiers, "google:import-org-3")
			assert.Contains(t, identifiers, "github:import-org-3")

			identifiers = actual.Credentials[identity.CredentialsTypeSAML].Identifiers
			assert.Len(t, identifiers, 2)
			assert.Contains(t, identifiers, "okta:import-saml-org-3")
			assert.Contains(t, identifiers, "onelogin:import-saml-org-3")

			assert.Empty(t, []byte(gjson.GetBytes(actual.Credentials[identity.CredentialsTypePassword].Config, "hashed_password").String()))
		})

		t.Run("with hashed passwords", func(t *testing.T) {
			for i, tt := range []struct{ name, hash, pass string }{
				{
					name: "pkbdf2",
					hash: "$pbkdf2-sha256$i=1000,l=128$e8/arsEf4cvQihdNgqj0Nw$5xQQKNTyeTHx2Ld5/JDE7A",
					pass: "123456",
				}, {
					name: "bcrypt2",
					hash: "$2a$10$ZsCsoVQ3xfBG/K2z2XpBf.tm90GZmtOqtqWcB5.pYd5Eq8y7RlDyq",
					pass: "123456",
				}, {
					name: "argon2i",
					hash: "$argon2i$v=19$m=65536,t=3,p=4$STVE4CQ9qQ1dK/j224VMbA$o8b+k5wdHgBqf7ES+aWG2K7Y9diQ6ahEhbW8zcstXGo",
					pass: "123456",
				}, {
					name: "argon2id",
					hash: "$argon2id$v=19$m=16,t=2,p=1$bVI1aE1SaTV6SGQ3bzdXdw$fnjCcZYmEPOUOjYXsT92Cg",
					pass: "123456",
				}, {
					name: "scrypt",
					hash: "$scrypt$ln=16384,r=8,p=1$ZtQva9xCHzlSELH/mA7Kj5KjH2tCrkbwYzdxknkL0QQ=$pnTcXKaWVT+FwFDdk3vO1K0J7ZgOxdSU1tCJNYmn8zI=",
					pass: "123456",
				}, {
					name: "md5",
					hash: "$md5$4QrcOUm6Wau+VuBX8g+IPg==",
					pass: "123456",
				}, {
					name: "SSHA",
					hash: "{SSHA}JFZFs0oHzxbMwkSJmYVeI8MnTDy/276a",
					pass: "test123",
				}, {
					name: "SSHA256",
					hash: "{SSHA256}czO44OTV17PcF1cRxWrLZLy9xHd7CWyVYplr1rOhuMlx/7IK",
					pass: "test123",
				}, {
					name: "SSHA512",
					hash: "{SSHA512}xPUl/px+1cG55rUH4rzcwxdOIPSB2TingLpiJJumN2xyDWN4Ix1WQG3ihnvHaWUE8MYNkvMi5rf0C9NYixHsE6Yh59M=",
					pass: "test123",
				}, {
					name: "hmac",
					hash: "$hmac-sha256$YjhhZDA4YTNhNTQ3ZTM1ODI5YjgyMWI3NTM3MDMwMWRkOGM0YjA2YmRkNzc3MWY5YjU0MWE3NTkxNDA2ODcxOA==$MTIzNDU2",
					pass: "123456",
				},
			} {
				t.Run("hash="+tt.name, func(t *testing.T) {
					traits := fmt.Sprintf(`{"email": "import-hash-%d@ory.sh"}`, i)
					res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.CreateIdentityBody{
						Traits: []byte(traits),
						Credentials: &identity.IdentityWithCredentials{Password: &identity.AdminIdentityImportCredentialsPassword{
							Config: identity.AdminIdentityImportCredentialsPasswordConfig{HashedPassword: tt.hash},
						}},
					})
					actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
					require.NoError(t, err)

					snapshotx.SnapshotT(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), snapshotx.ExceptNestedKeys(ignoreDefault...), snapshotx.ExceptNestedKeys("hashed_password"))

					require.NoError(t, hash.Compare(ctx, []byte(tt.pass), []byte(gjson.GetBytes(actual.Credentials[identity.CredentialsTypePassword].Config, "hashed_password").String())))
				})
			}
		})

		t.Run("with password migration hook enabled", func(t *testing.T) {
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.CreateIdentityBody{
				Traits: []byte(`{"email": "pw-migration-hook@ory.sh"}`),
				Credentials: &identity.IdentityWithCredentials{Password: &identity.AdminIdentityImportCredentialsPassword{
					Config: identity.AdminIdentityImportCredentialsPasswordConfig{UsePasswordMigrationHook: true},
				}},
			})
			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			snapshotx.SnapshotT(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), snapshotx.ExceptNestedKeys(ignoreDefault...), snapshotx.ExceptNestedKeys("hashed_password"))

			assert.True(t, gjson.GetBytes(actual.Credentials[identity.CredentialsTypePassword].Config, "use_password_migration_hook").Bool())
		})

		t.Run("with not-normalized email", func(t *testing.T) {
			res := send(t, adminTS, "POST", "/identities", http.StatusCreated, identity.CreateIdentityBody{
				SchemaID: "customer",
				Traits:   []byte(`{"email": "UpperCased@ory.sh"}`),
				VerifiableAddresses: []identity.VerifiableAddress{{
					Verified: true,
					Value:    "UpperCased@ory.sh",
					Via:      identity.VerifiableAddressTypeEmail,
					Status:   identity.VerifiableAddressStatusCompleted,
				}},
				RecoveryAddresses: []identity.RecoveryAddress{{Value: "UpperCased@ory.sh"}},
			})
			actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
			require.NoError(t, err)

			require.Len(t, actual.VerifiableAddresses, 1)
			assert.True(t, actual.VerifiableAddresses[0].Verified)
			assert.Equal(t, "uppercased@ory.sh", actual.VerifiableAddresses[0].Value)

			require.Len(t, actual.RecoveryAddresses, 1)
			assert.Equal(t, "uppercased@ory.sh", actual.RecoveryAddresses[0].Value)

			snapshotx.SnapshotT(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), snapshotx.ExceptNestedKeys(ignoreDefault...), snapshotx.ExceptNestedKeys("verified_at"))
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

	t.Run("suite=create and batch list", func(t *testing.T) {
		var ids []uuid.UUID
		identitiesAmount := 5
		listAmount := 3
		t.Run("case= create multiple identities", func(t *testing.T) {
			for i := 0; i < identitiesAmount; i++ {
				res := send(t, adminTS, "POST", "/identities", http.StatusCreated, json.RawMessage(`{"traits": {"bar":"baz"}}`))
				assert.NotEmpty(t, res.Get("id").String(), "%s", res.Raw)

				id := x.ParseUUID(res.Get("id").String())
				ids = append(ids, id)
			}
			require.Len(t, ids, identitiesAmount)
		})

		t.Run("case=list few identities", func(t *testing.T) {
			vals := url.Values{}
			vals.Add("ids", ids[0].String()) // duplicate ID is deduplicated in result
			for i := range listAmount {
				vals.Add("ids", ids[i].String())
			}
			res := get(t, adminTS, "/identities?"+vals.Encode(), http.StatusOK)

			identities := res.Array()
			require.Len(t, identities, listAmount)
		})
	})

	t.Run("case=list identities by ID is capped at 500", func(t *testing.T) {
		vals := url.Values{}
		for range 501 {
			vals.Add("ids", x.NewUUID().String())
		}
		res := get(t, adminTS, "/identities?"+vals.Encode(), http.StatusBadRequest)
		assert.Contains(t, res.Get("error.reason").String(), "must not exceed 500")
	})

	t.Run("case=list identities cannot combine filters", func(t *testing.T) {
		filters := []string{
			"ids=" + x.NewUUID().String(),
			"credentials_identifier=foo@bar.com",
			"preview_credentials_identifier_similar=bar.com",
			"organization_id=" + x.NewUUID().String(),
		}
		for i := range filters {
			for j := range filters {
				if i == j {
					continue // OK to use the same filter multiple times. Behavior varies by filter, though.
				}

				u := "/identities?" + filters[i] + "&" + filters[j]
				res := get(t, adminTS, u, http.StatusBadRequest)
				assert.Contains(t, res.Get("error.reason").String(), "cannot combine multiple filters")
			}
		}
	})

	t.Run("case=malformed ids should return an error", func(t *testing.T) {
		res := get(t, adminTS, "/identities?ids=not-a-uuid", http.StatusBadRequest)
		assert.Contains(t, res.Get("error.reason").String(), "Invalid UUID value `not-a-uuid` for parameter `ids`.", "%s", res.Raw)
	})

	t.Run("suite=create and update", func(t *testing.T) {
		var i identity.Identity
		createOidcIdentity := func(t *testing.T, identifier, accessToken, refreshToken, idToken string, encrypt bool) string {
			transform := func(token, suffix string) string {
				if !encrypt {
					return token
				}
				if token == "" {
					return ""
				}
				c, err := reg.Cipher(ctx).Encrypt(context.Background(), []byte(token+suffix))
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
								InitialAccessToken:  transform(accessToken, "0"),
								InitialRefreshToken: transform(refreshToken, "0"),
								InitialIDToken:      transform(idToken, "0"),
							},
							{
								Subject:             "baz",
								Provider:            "zab",
								InitialAccessToken:  transform(accessToken, "1"),
								InitialRefreshToken: transform(refreshToken, "1"),
								InitialIDToken:      transform(idToken, "1"),
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

		t.Run("case=should return an empty array on a failed lookup with identifier", func(t *testing.T) {
			res := get(t, adminTS, "/identities?credentials_identifier=find.by.non.existing.identifier@bar.com", http.StatusOK)
			assert.EqualValues(t, int64(0), res.Get("#").Int(), "%s", res.Raw)
		})

		t.Run("case=should be able to lookup the identity using identifier", func(t *testing.T) {
			ident := &identity.Identity{
				Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypePassword: {
						Type:        identity.CredentialsTypePassword,
						Identifiers: []string{"find.by.identifier@bar.com"},
						Config:      sqlxx.JSONRawMessage(`{"hashed_password":"$2a$08$.cOYmAd.vCpDOoiVJrO5B.hjTLKQQ6cAK40u8uB.FnZDyPvVvQ9Q."}`), // foobar
					},
					identity.CredentialsTypeOIDC: {
						Type:        identity.CredentialsTypeOIDC,
						Identifiers: []string{"ProviderID:293b5d9b-1009-4600-a3e9-bd1845de22f2"},
						Config:      sqlxx.JSONRawMessage("{\"some\" : \"secret\"}"),
					},
				},
				State:  identity.StateActive,
				Traits: identity.Traits(`{"username":"find.by.identifier@bar.com"}`),
			}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), ident))

			t.Run("type=password", func(t *testing.T) {
				res := get(t, adminTS, "/identities?credentials_identifier=FIND.BY.IDENTIFIER@bar.com", http.StatusOK)
				assert.EqualValues(t, ident.ID.String(), res.Get("0.id").String(), "%s", res.Raw)
				assert.EqualValues(t, "find.by.identifier@bar.com", res.Get("0.traits.username").String(), "%s", res.Raw)
				assert.EqualValues(t, defaultSchemaExternalURL, res.Get("0.schema_url").String(), "%s", res.Raw)
				assert.EqualValues(t, config.DefaultIdentityTraitsSchemaID, res.Get("0.schema_id").String(), "%s", res.Raw)
				assert.EqualValues(t, identity.StateActive, res.Get("0.state").String(), "%s", res.Raw)
				assert.EqualValues(t, "password", res.Get("0.credentials.password.type").String(), res.Raw)
				assert.EqualValues(t, "1", res.Get("0.credentials.password.identifiers.#").String(), res.Raw)
				assert.EqualValues(t, "find.by.identifier@bar.com", res.Get("0.credentials.password.identifiers.0").String(), res.Raw)
			})

			t.Run("type=oidc", func(t *testing.T) {
				res := get(t, adminTS, "/identities?credentials_identifier=ProviderID:293b5d9b-1009-4600-a3e9-bd1845de22f2", http.StatusOK)
				assert.EqualValues(t, ident.ID.String(), res.Get("0.id").String(), "%s", res.Raw)
				assert.EqualValues(t, "find.by.identifier@bar.com", res.Get("0.traits.username").String(), "%s", res.Raw)
				assert.EqualValues(t, defaultSchemaExternalURL, res.Get("0.schema_url").String(), "%s", res.Raw)
				assert.EqualValues(t, config.DefaultIdentityTraitsSchemaID, res.Get("0.schema_id").String(), "%s", res.Raw)
				assert.EqualValues(t, identity.StateActive, res.Get("0.state").String(), "%s", res.Raw)
				assert.EqualValues(t, "oidc", res.Get("0.credentials.oidc.type").String(), res.Raw)
				assert.EqualValues(t, "1", res.Get("0.credentials.oidc.identifiers.#").String(), res.Raw)
				assert.EqualValues(t, "ProviderID:293b5d9b-1009-4600-a3e9-bd1845de22f2", res.Get("0.credentials.oidc.identifiers.0").String(), res.Raw)
			})
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

		t.Run("case=should not fail on empty tokens", func(t *testing.T) {
			id := createOidcIdentity(t, "foo.oidc.empty-tokens@bar.com", "", "", "", true)
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
					assert.EqualValues(t, "", res.Get("credentials.oidc.config.providers.0.initial_access_token").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "", res.Get("credentials.oidc.config.providers.0.initial_refresh_token").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "", res.Get("credentials.oidc.config.providers.0.initial_id_token").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "baz", res.Get("credentials.oidc.config.providers.1.subject").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "zab", res.Get("credentials.oidc.config.providers.1.provider").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "", res.Get("credentials.oidc.config.providers.1.initial_access_token").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "", res.Get("credentials.oidc.config.providers.1.initial_refresh_token").String(), "credentials should be included: %s", res.Raw)
					assert.EqualValues(t, "", res.Get("credentials.oidc.config.providers.1.initial_id_token").String(), "credentials should be included: %s", res.Raw)
				})
			}
		})

		t.Run("case=should get identity with credentials", func(t *testing.T) {
			i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			credentials := map[identity.CredentialsType]identity.Credentials{
				identity.CredentialsTypePassword: {Identifiers: []string{"zab", "bar"}, Type: identity.CredentialsTypePassword, Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\"}")},
				identity.CredentialsTypeOIDC:     {Type: identity.CredentialsTypeOIDC, Identifiers: []string{"bar", "baz"}, Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\"}")},
				identity.CredentialsTypeWebAuthn: {Type: identity.CredentialsTypeWebAuthn, Identifiers: []string{"foo", "bar"}, Config: sqlxx.JSONRawMessage("{\"some\" : \"secret\", \"user_handle\": \"rVIFaWRcTTuQLkXFmQWpgA==\"}")},
			}
			i.Credentials = credentials
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

			excludeKeys := snapshotx.ExceptNestedKeys("id", "created_at", "updated_at", "schema_url", "state_changed_at")
			t.Run("case=should get identity without credentials included", func(t *testing.T) {
				res := get(t, adminTS, "/identities/"+i.ID.String(), http.StatusOK)
				snapshotx.SnapshotT(t, json.RawMessage(res.Raw), excludeKeys)
			})

			t.Run("case=should get identity with password credentials included", func(t *testing.T) {
				res := get(t, adminTS, "/identities/"+i.ID.String()+"?include_credential=password", http.StatusOK)
				snapshotx.SnapshotT(t, json.RawMessage(res.Raw), excludeKeys)
			})

			t.Run("case=should get identity with password and webauthn credentials included", func(t *testing.T) {
				res := get(t, adminTS, "/identities/"+i.ID.String()+"?include_credential=password&include_credential=webauthn", http.StatusOK)
				snapshotx.SnapshotT(t, json.RawMessage(res.Raw), excludeKeys)
			})
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

		t.Run("case=should return empty tokens if decryption fails", func(t *testing.T) {
			id := createOidcIdentity(t, "foo-failed.oidc@bar.com", "foo_token", "bar_token", "id_token", false)
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					res := get(t, ts, "/identities/"+i.ID.String()+"?include_credential=oidc", http.StatusOK)
					assert.NotContains(t, res.Raw, "identifier_credentials", res.Raw)

					res = get(t, ts, "/identities/"+id+"?include_credential=oidc", http.StatusOK)
					assert.Equal(t, "bar:foo-failed.oidc@bar.com", res.Get("credentials.oidc.identifiers.0").String(), "%s", res.Raw)
					assert.Equal(t, "", res.Get("credentials.oidc.config.providers.0.initial_access_token").String(), "%s", res.Raw)
					assert.Equal(t, "", res.Get("credentials.oidc.config.providers.0.initial_id_token").String(), "%s", res.Raw)
					assert.Equal(t, "", res.Get("credentials.oidc.config.providers.0.initial_refresh_token").String(), "%s", res.Raw)
				})
			}
		})

		t.Run("case=should return decrypted token", func(t *testing.T) {
			e, _ := reg.Cipher(ctx).Encrypt(context.Background(), []byte("foo_token"))
			id := createOidcIdentity(t, "foo-failed-2.oidc@bar.com", e, "bar_token", "id_token", false)
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					t.Logf("no oidc token")
					res := get(t, ts, "/identities/"+i.ID.String()+"?include_credential=oidc", http.StatusOK)
					assert.NotContains(t, res.Raw, "identifier_credentials", res.Raw)

					t.Logf("get oidc token")
					res = get(t, ts, "/identities/"+id+"?include_credential=oidc", http.StatusOK)
					assert.Equal(t, "bar:foo-failed-2.oidc@bar.com", res.Get("credentials.oidc.identifiers.0").String(), "%s", res.Raw)
					assert.Equal(t, "foo_token", res.Get("credentials.oidc.config.providers.0.initial_access_token").String(), "%s", res.Raw)
				})
			}
		})

		t.Run("case=should update an identity and persist the changes", func(t *testing.T) {
			i := &identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, x.NewUUID().String()))}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					ur := identity.UpdateIdentityBody{
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
		t.Run("case=should update an identity with credentials", func(t *testing.T) {
			i := &identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, x.NewUUID().String()))}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					credentials := identity.IdentityWithCredentials{
						Password: &identity.AdminIdentityImportCredentialsPassword{
							Config: identity.AdminIdentityImportCredentialsPasswordConfig{
								Password: "pswd1234",
							},
						},
					}
					ur := identity.UpdateIdentityBody{
						Traits:         []byte(`{"bar":"baz","foo":"baz"}`),
						SchemaID:       i.SchemaID,
						State:          identity.StateInactive,
						MetadataPublic: []byte(`{"public":"metadata"}`),
						MetadataAdmin:  []byte(`{"admin":"metadata"}`),
						Credentials:    &credentials,
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
					actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(context.Background(), i.ID)
					require.NoError(t, err)
					require.NoError(t, hash.Compare(ctx, []byte("pswd1234"), []byte(gjson.GetBytes(actual.Credentials[identity.CredentialsTypePassword].Config, "hashed_password").String())))
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

		t.Run("case=should create an identity with linking marker", func(t *testing.T) {
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					trait := x.NewUUID().String()
					payload := `
						{
							"traits": {
								"bar": "` + trait + `"
							},
							"credentials": {
								"oidc": {
									"config": {
										"providers": [
											{
												"subject": "` + trait + `",
												"provider": "bar",
												"use_auto_link": true
											}
										]
									}
								}
							}
						}`

					res := send(t, ts, "POST", "/identities", http.StatusCreated, json.RawMessage(payload))
					i.ID = x.ParseUUID(res.Get("id").String())

					identRes := send(t, adminTS, "GET", fmt.Sprintf("/identities/%s?include_credential=oidc", i.ID), http.StatusOK, nil)

					assert.True(t, identRes.Get("credentials.oidc.config.providers.0.use_auto_link").Bool())
					assert.False(t, identRes.Get("credentials.oidc.config.providers.0.organization").Exists())
				})
			}
		})

		t.Run("case=should create an identity without linking marker omitempty", func(t *testing.T) {
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					trait := x.NewUUID().String()
					payload := `
						{
							"traits": {
								"bar": "` + trait + `"
							},
							"credentials": {
								"oidc": {
									"config": {
										"providers": [
											{
												"subject": "` + trait + `",
												"provider": "bar",
												"use_auto_link": false
											}
										]
									}
								}
							}
						}`
					res := send(t, ts, "POST", "/identities", http.StatusCreated, json.RawMessage(payload))
					stateChangedAt := sqlxx.NullTime(res.Get("state_changed_at").Time())

					i.Traits = []byte(res.Get("traits").Raw)
					i.ID = x.ParseUUID(res.Get("id").String())
					i.StateChangedAt = &stateChangedAt
					assert.NotEmpty(t, res.Get("id").String())

					i, err := reg.Persister().GetIdentityConfidential(context.Background(), i.ID)
					require.NoError(t, err)

					require.False(t, gjson.GetBytes(i.Credentials[identity.CredentialsTypeOIDC].Config, "providers.0.use_auto_link").Exists())
				})
			}
		})
	})

	t.Run("suite=PATCH identities", func(t *testing.T) {
		t.Run("case=fails with too many patches", func(t *testing.T) {
			tooMany := make([]*identity.BatchIdentityPatch, identity.BatchPatchIdentitiesLimit+1)
			for i := range tooMany {
				tooMany[i] = &identity.BatchIdentityPatch{Create: validCreateIdentityBody("too-many-patches", i)}
			}
			res := send(t, adminTS, "PATCH", "/identities", http.StatusBadRequest,
				&identity.BatchPatchIdentitiesBody{Identities: tooMany})
			assert.Contains(t, res.Get("error.reason").String(), strconv.Itoa(identity.BatchPatchIdentitiesLimit),
				"the error reason should contain the limit")
		})
		t.Run("case=fails some on a bad identity", func(t *testing.T) {
			// Test setup: we have a list of valid identitiy patches and a list of invalid ones.
			// Each run adds one invalid patch to the list and sends it to the server.
			// --> we expect the server to fail only the bad patches in the list.
			// Finally, we send just valid patches
			// --> we expect the server to succeed all patches in the list.

			t.Run("case=invalid patches fail", func(t *testing.T) {
				patches := []*identity.BatchIdentityPatch{
					{Create: validCreateIdentityBody("valid", 0)},
					{Create: validCreateIdentityBody("valid", 1)},
					{Create: &identity.CreateIdentityBody{}}, // <-- invalid: missing all fields
					{Create: validCreateIdentityBody("valid", 2)},
					{Create: validCreateIdentityBody("valid", 0)}, // <-- duplicate
					{Create: validCreateIdentityBody("valid", 3)},
					{Create: &identity.CreateIdentityBody{Traits: json.RawMessage(`"invalid traits"`)}}, // <-- invalid traits
					{Create: validCreateIdentityBody("valid", 4)},
				}
				expectedToPass := []*identity.BatchIdentityPatch{patches[0], patches[1], patches[3], patches[5], patches[7]}

				// Create unique IDs for each patch
				patchIDs := make([]string, len(patches))
				for i, p := range patches {
					id := uuid.NewV5(uuid.Nil, fmt.Sprintf("%d", i))
					p.ID = &id
					patchIDs[i] = id.String()
				}

				req := &identity.BatchPatchIdentitiesBody{Identities: patches}
				body := send(t, adminTS, "PATCH", "/identities", http.StatusOK, req)
				var actions []string
				require.NoErrorf(t, json.Unmarshal(([]byte)(body.Get("identities.#.action").Raw), &actions), "%s", body)
				assert.Equalf(t,
					[]string{"create", "create", "error", "create", "error", "create", "error", "create"},
					actions, "%s", body)

				// Check that all patch IDs are returned
				for i, gotPatchID := range body.Get("identities.#.patch_id").Array() {
					assert.Equal(t, patchIDs[i], gotPatchID.String())
				}

				// Check specific errors
				assert.Equal(t, "Bad Request", body.Get("identities.2.error.status").String())
				assert.Equal(t, "Conflict", body.Get("identities.4.error.status").String())
				assert.Equal(t, "Bad Request", body.Get("identities.6.error.status").String())

				var identityIDs []uuid.UUID
				require.NoErrorf(t, json.Unmarshal(([]byte)(body.Get("identities.#.identity").Raw), &identityIDs), "%s", body)

				actualIdentities, _, err := reg.Persister().ListIdentities(ctx, identity.ListIdentityParameters{IdsFilter: identityIDs})
				require.NoError(t, err)
				actualIdentityIDs := make([]uuid.UUID, len(actualIdentities))
				for i, id := range actualIdentities {
					actualIdentityIDs[i] = id.ID
				}
				assert.ElementsMatchf(t, identityIDs, actualIdentityIDs, "%s", body)

				expectedTraits := make(map[string]string, len(expectedToPass))
				for i, p := range expectedToPass {
					expectedTraits[identityIDs[i].String()] = string(p.Create.Traits)
				}
				actualTraits := make(map[string]string, len(actualIdentities))
				for _, id := range actualIdentities {
					actualTraits[id.ID.String()] = string(id.Traits)
				}

				assert.Equal(t, expectedTraits, actualTraits)
			})

			t.Run("valid patches succeed", func(t *testing.T) {
				validPatches := []*identity.BatchIdentityPatch{
					{Create: validCreateIdentityBody("valid-patch", 0)},
					{Create: validCreateIdentityBody("valid-patch", 1)},
					{Create: validCreateIdentityBody("valid-patch", 2)},
					{Create: validCreateIdentityBody("valid-patch", 3)},
					{Create: validCreateIdentityBody("valid-patch", 4)},
				}
				req := &identity.BatchPatchIdentitiesBody{Identities: validPatches}
				send(t, adminTS, "PATCH", "/identities", http.StatusOK, req)
			})
		})

		t.Run("case=ignores create nil bodies", func(t *testing.T) {
			patches := []*identity.BatchIdentityPatch{
				{Create: nil},
				{Create: validCreateIdentityBody("nil-batch-import", 0)},
				{Create: nil},
				{Create: validCreateIdentityBody("nil-batch-import", 1)},
				{Create: nil},
				{Create: validCreateIdentityBody("nil-batch-import", 2)},
				{Create: nil},
				{Create: validCreateIdentityBody("nil-batch-import", 3)},
				{Create: nil},
			}
			req := &identity.BatchPatchIdentitiesBody{Identities: patches}
			res := send(t, adminTS, "PATCH", "/identities", http.StatusOK, req)
			assert.Len(t, res.Get("identities").Array(), len(patches))
			assert.Equal(t, "null", res.Get("identities.0").Raw)
			assert.Equal(t, "null", res.Get("identities.2").Raw)
			assert.Equal(t, "null", res.Get("identities.4").Raw)
			assert.Equal(t, "null", res.Get("identities.6").Raw)
			assert.Equal(t, "null", res.Get("identities.8").Raw)
		})

		t.Run("case=success", func(t *testing.T) {
			patches := []*identity.BatchIdentityPatch{
				{Create: validCreateIdentityBody("Batch-Import", 0)},
				{Create: validCreateIdentityBody("batch-import", 1)},
				{Create: validCreateIdentityBody("batch-import", 2)},
				{Create: validCreateIdentityBody("batch-import", 3)},
			}
			req := &identity.BatchPatchIdentitiesBody{Identities: patches}
			res := send(t, adminTS, "PATCH", "/identities", http.StatusOK, req)

			assert.Len(t, res.Get("identities").Array(), len(patches))

			for i, patch := range patches {
				t.Run(fmt.Sprintf("assert=identity %d", i), func(t *testing.T) {
					identityID := res.Get(fmt.Sprintf("identities.%d.identity", i)).String()
					require.NotEmpty(t, identityID)

					res := get(t, adminTS, "/identities/"+identityID, http.StatusOK)
					snapshotx.SnapshotT(t, res.Value(), snapshotx.ExceptNestedKeys(
						// All these keys change randomly, so we need to test them individually below
						"id", "schema_url",
						"created_at", "updated_at", "state_changed_at",
						"verifiable_addresses", "recovery_addresses", "identifiers"))

					emails := gjson.Parse(strings.ToLower(gjson.GetBytes(patch.Create.Traits, "emails").Raw))
					assert.Equal(t, identityID, res.Get("id").String())
					assert.EqualValues(t, patch.Create.Traits, res.Get("traits").Raw)
					assertJSONArrayElementsMatch(t, emails, res.Get("credentials.password.identifiers"))
					assertJSONArrayElementsMatch(t, emails, res.Get("recovery_addresses.#.value"))
					assertJSONArrayElementsMatch(t, emails, res.Get("verifiable_addresses.#.value"))

					// Test that the verified addresses are imported correctly
					assert.Len(t, res.Get("verifiable_addresses.#(verified=true)#").Array(), 2)
					assert.Len(t, res.Get("verifiable_addresses.#(verified=false)#").Array(), 2)
					assert.Len(t, res.Get("verifiable_addresses.#(status=pending)#").Array(), 2)
					assert.Len(t, res.Get("verifiable_addresses.#(status=sent)#").Array(), 1)
					assert.Len(t, res.Get("verifiable_addresses.#(status=completed)#").Array(), 1)
				})
			}
		})
	})

	t.Run("case=PATCH update of state should update state changed at timestamp", func(t *testing.T) {
		id := x.NewUUID().String()
		email := "UPPER" + id + "@ory.sh"
		i := &identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"subject": %q, "email": %q}`, id, email))}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				patch := []patch{
					{"op": "replace", "path": "/state", "value": identity.StateInactive},
				}

				res := send(t, ts, "PATCH", "/identities/"+i.ID.String(), http.StatusOK, &patch)
				assert.EqualValues(t, id, res.Get("traits.subject").String(), "%s", res.Raw)
				assert.EqualValues(t, email, res.Get("traits.email").String(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_admin.admin").Exists(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_public.public").Exists(), "%s", res.Raw)
				assert.EqualValues(t, identity.StateInactive, res.Get("state").String(), "%s", res.Raw)
				assert.NotEqualValues(t, i.StateChangedAt, sqlxx.NullTime(res.Get("state_changed_at").Time()), "%s", res.Raw)

				res = get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
				assert.EqualValues(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)
				assert.EqualValues(t, id, res.Get("traits.subject").String(), "%s", res.Raw)
				assert.EqualValues(t, email, res.Get("traits.email").String(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_admin.admin").Exists(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_public.public").Exists(), "%s", res.Raw)
				assert.EqualValues(t, identity.StateInactive, res.Get("state").String(), "%s", res.Raw)
				assert.NotEqualValues(t, i.StateChangedAt, sqlxx.NullTime(res.Get("state_changed_at").Time()), "%s", res.Raw)
			})
		}
	})

	t.Run("case=PATCH update with uppercase emails should work", func(t *testing.T) {
		// Regression test for https://github.com/ory/kratos/issues/3187

		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				email := "UPPER" + x.NewUUID().String() + "@ory.sh"
				lowercaseEmail := strings.ToLower(email)
				var cr identity.CreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + email + `"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)
				assert.EqualValues(t, lowercaseEmail, res.Get("recovery_addresses.0.value").String(), "%s", res.Raw)
				assert.EqualValues(t, lowercaseEmail, res.Get("verifiable_addresses.0.value").String(), "%s", res.Raw)
				identityID := res.Get("id").String()

				patch := []patch{
					{
						"op":    "replace",
						"path":  "/verifiable_addresses/0/verified",
						"value": true,
					},
				}

				res = send(t, ts, "PATCH", "/identities/"+identityID, http.StatusOK, &patch)
				assert.EqualValues(t, email, res.Get("traits.email").String(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_admin.admin").Exists(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_public.public").Exists(), "%s", res.Raw)
				assert.EqualValues(t, identity.StateActive, res.Get("state").String(), "%s", res.Raw)

				res = get(t, ts, "/identities/"+identityID, http.StatusOK)
				assert.EqualValues(t, identityID, res.Get("id").String(), "%s", res.Raw)
				assert.EqualValues(t, email, res.Get("traits.email").String(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_admin.admin").Exists(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_public.public").Exists(), "%s", res.Raw)
				assert.EqualValues(t, identity.StateActive, res.Get("state").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=PATCH should update verified_at timestamp", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				email := x.NewUUID().String() + "@ory.sh"
				var cr identity.CreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + email + `"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)
				assert.EqualValues(t, email, res.Get("recovery_addresses.0.value").String(), "%s", res.Raw)
				assert.EqualValues(t, email, res.Get("verifiable_addresses.0.value").String(), "%s", res.Raw)
				assert.Falsef(t, res.Get("verifiable_addresses.0.verified").Bool(), "%s", res.Raw)
				assert.Falsef(t, res.Get("verifiable_addresses.0.verified_at").Exists(), "%s", res.Raw)
				identityID := res.Get("id").String()

				// set to verified, should also update verified_at timestamp
				patch1 := []patch{
					{
						"op":    "replace",
						"path":  "/verifiable_addresses/0/verified",
						"value": true,
					},
				}

				now := time.Now()

				res = send(t, ts, "PATCH", "/identities/"+identityID, http.StatusOK, &patch1)
				assert.EqualValues(t, email, res.Get("recovery_addresses.0.value").String(), "%s", res.Raw)
				assert.EqualValues(t, email, res.Get("verifiable_addresses.0.value").String(), "%s", res.Raw)
				assert.Truef(t, res.Get("verifiable_addresses.0.verified").Bool(), "%s", res.Raw)
				assert.WithinDurationf(t, now, res.Get("verifiable_addresses.0.updated_at").Time(), 5*time.Second, "%s", res.Raw)
				assert.WithinDurationf(t, now, res.Get("verifiable_addresses.0.verified_at").Time(), 5*time.Second, "%s", res.Raw)

				res = get(t, ts, "/identities/"+identityID, http.StatusOK)
				assert.EqualValues(t, email, res.Get("recovery_addresses.0.value").String(), "%s", res.Raw)
				assert.EqualValues(t, email, res.Get("verifiable_addresses.0.value").String(), "%s", res.Raw)
				assert.Truef(t, res.Get("verifiable_addresses.0.verified").Bool(), "%s", res.Raw)
				assert.WithinDurationf(t, now, res.Get("verifiable_addresses.0.updated_at").Time(), 5*time.Second, "%s", res.Raw)
				assert.WithinDurationf(t, now, res.Get("verifiable_addresses.0.verified_at").Time(), 5*time.Second, "%s", res.Raw)

				// update only verified_at timestamp
				verifiedAt := time.Date(1999, 1, 7, 8, 23, 19, 0, time.UTC)
				patch2 := []patch{
					{
						"op":    "replace",
						"path":  "/verifiable_addresses/0/verified_at",
						"value": verifiedAt.Format(time.RFC3339),
					},
				}

				now = time.Now()
				res = send(t, ts, "PATCH", "/identities/"+identityID, http.StatusOK, &patch2)
				assert.EqualValues(t, email, res.Get("recovery_addresses.0.value").String(), "%s", res.Raw)
				assert.EqualValues(t, email, res.Get("verifiable_addresses.0.value").String(), "%s", res.Raw)
				assert.Truef(t, res.Get("verifiable_addresses.0.verified").Bool(), "%s", res.Raw)
				assert.Equalf(t, verifiedAt, res.Get("verifiable_addresses.0.verified_at").Time(), "%s", res.Raw)
				assert.WithinDurationf(t, now, res.Get("verifiable_addresses.0.updated_at").Time(), 5*time.Second, "%s", res.Raw)

				res = get(t, ts, "/identities/"+identityID, http.StatusOK)
				assert.EqualValues(t, email, res.Get("recovery_addresses.0.value").String(), "%s", res.Raw)
				assert.EqualValues(t, email, res.Get("verifiable_addresses.0.value").String(), "%s", res.Raw)
				assert.Truef(t, res.Get("verifiable_addresses.0.verified").Bool(), "%s", res.Raw)
				assert.Equalf(t, verifiedAt, res.Get("verifiable_addresses.0.verified_at").Time(), "%s", res.Raw)
				assert.WithinDurationf(t, now, res.Get("verifiable_addresses.0.updated_at").Time(), 5*time.Second, "%s", res.Raw)

				// remove verified status
				patch3 := []patch{
					{
						"op":    "replace",
						"path":  "/verifiable_addresses/0/verified",
						"value": false,
					},
				}

				now = time.Now()

				res = send(t, ts, "PATCH", "/identities/"+identityID, http.StatusOK, &patch3)
				assert.EqualValues(t, email, res.Get("recovery_addresses.0.value").String(), "%s", res.Raw)
				assert.EqualValues(t, email, res.Get("verifiable_addresses.0.value").String(), "%s", res.Raw)
				assert.Falsef(t, res.Get("verifiable_addresses.0.verified").Bool(), "%s", res.Raw)
				assert.Falsef(t, res.Get("verifiable_addresses.0.verified_at").Exists(), "%s", res.Raw)
				assert.WithinDurationf(t, now, res.Get("verifiable_addresses.0.updated_at").Time(), 5*time.Second, "%s", res.Raw)

				res = get(t, ts, "/identities/"+identityID, http.StatusOK)
				assert.EqualValues(t, email, res.Get("recovery_addresses.0.value").String(), "%s", res.Raw)
				assert.EqualValues(t, email, res.Get("verifiable_addresses.0.value").String(), "%s", res.Raw)
				assert.Falsef(t, res.Get("verifiable_addresses.0.verified").Bool(), "%s", res.Raw)
				assert.Falsef(t, res.Get("verifiable_addresses.0.verified_at").Exists(), "%s", res.Raw)
				assert.WithinDurationf(t, now, res.Get("verifiable_addresses.0.updated_at").Time(), 5*time.Second, "%s", res.Raw)
			})
		}
	})

	t.Run("case=PATCH update should not persist if schema id is invalid", func(t *testing.T) {
		sub := x.NewUUID().String()
		i := &identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, sub))}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				patch := []patch{
					{"op": "replace", "path": "/schema_id", "value": "invalid-id"},
				}

				res := send(t, ts, "PATCH", "/identities/"+i.ID.String(), http.StatusBadRequest, &patch)
				assert.Contains(t, res.Get("error.reason").String(), "invalid-id", "%s", res.Raw)

				res = get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
				// Assert that the schema ID is unchanged
				assert.EqualValues(t, i.SchemaID, res.Get("schema_id").String(), "%s", res.Raw)
				assert.EqualValues(t, sub, res.Get("traits.subject").String(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_admin.admin").Exists(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_public.public").Exists(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=PATCH update should not persist if invalid state is supplied", func(t *testing.T) {
		sub := x.NewUUID().String()
		i := &identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, sub))}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				patch := []patch{
					{"op": "replace", "path": "/state", "value": "invalid-value"},
				}

				res := send(t, ts, "PATCH", "/identities/"+i.ID.String(), http.StatusBadRequest, &patch)
				assert.EqualValues(t, "The supplied state ('invalid-value') was not valid. Valid states are ('active', 'inactive').", res.Get("error.reason").String(), "%s", res.Raw)

				res = get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
				// Assert that the schema ID is unchanged
				assert.EqualValues(t, i.SchemaID, res.Get("schema_id").String(), "%s", res.Raw)
				assert.EqualValues(t, sub, res.Get("traits.subject").String(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_admin.admin").Exists(), "%s", res.Raw)
				assert.False(t, res.Get("metadata_public.public").Exists(), "%s", res.Raw)
				assert.NotEqualValues(t, i.StateChangedAt, sqlxx.NullTime(res.Get("state_changed_at").Time()), "%s", res.Raw)
			})
		}
	})

	t.Run("case=PATCH update should update nested fields", func(t *testing.T) {
		sub := x.NewUUID().String()
		i := &identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, sub))}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				patch := []patch{
					{"op": "replace", "path": "/traits/subject", "value": "patched-subject"},
				}

				res := send(t, ts, "PATCH", "/identities/"+i.ID.String(), http.StatusOK, &patch)
				assert.EqualValues(t, "patched-subject", res.Get("traits.subject").String(), "%s", res.Raw)

				res = get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
				// Assert that the schema ID is unchanged
				assert.EqualValues(t, i.SchemaID, res.Get("schema_id").String(), "%s", res.Raw)
				assert.EqualValues(t, "patched-subject", res.Get("traits.subject").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=PATCH should fail if no JSON payload is sent", func(t *testing.T) {
		sub := x.NewUUID().String()
		i := &identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, sub))}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				res := send(t, ts, "PATCH", "/identities/"+i.ID.String(), http.StatusBadRequest, nil)
				assert.Equal(t, res.Get("error.message").Str, "invalid state detected", res.Raw)
			})
		}
	})

	t.Run("case=PATCH should fail if credentials are updated", func(t *testing.T) {
		sub := x.NewUUID().String()
		i := &identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, sub))}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				patch := []patch{
					{"op": "replace", "path": "/credentials", "value": "patched-credentials"},
				}

				res := send(t, ts, "PATCH", "/identities/"+i.ID.String(), http.StatusBadRequest, &patch)

				assert.EqualValues(t, "patch includes denied path: /credentials", res.Get("error.message").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=PATCH should fail if credential orgs are updated", func(t *testing.T) {
		email := x.NewUUID().String() + "@ory.sh"
		i := &identity.Identity{Traits: identity.Traits(`{"email":"` + email + `"}`)}
		i.SetCredentials(identity.CredentialsTypeOIDC, identity.Credentials{
			Type:        identity.CredentialsTypeOIDC,
			Identifiers: []string{email},
			Config:      sqlxx.JSONRawMessage(`{"providers": [{"provider": "some-provider"}]}`),
		})
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				patch := []patch{
					{"op": "replace", "path": "/credentials/oidc/config/providers/0/organization", "value": "foo"},
				}

				res := send(t, ts, "PATCH", "/identities/"+i.ID.String(), http.StatusBadRequest, &patch)

				assert.EqualValues(t, "patch includes denied path: /credentials/oidc/config/providers/0/organization", res.Get("error.message").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=PATCH should allow to update credential password", func(t *testing.T) {
		email := x.NewUUID().String() + "@ory.sh"
		password := "ljanf123akf"
		p, err := reg.Hasher(ctx).Generate(context.Background(), []byte(password))
		require.NoError(t, err)
		i := &identity.Identity{Traits: identity.Traits(`{"email":"` + email + `"}`)}
		i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
			Type:        identity.CredentialsTypePassword,
			Identifiers: []string{email},
			Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
		})
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				patch := []patch{
					{"op": "replace", "path": "/credentials/password/config/hashed_password", "value": "foo"},
				}

				send(t, ts, "PATCH", "/identities/"+i.ID.String(), http.StatusOK, &patch)

				updated, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, i.ID)
				require.NoError(t, err)
				assert.Equal(t, "foo",
					gjson.GetBytes(updated.Credentials[identity.CredentialsTypePassword].Config, "hashed_password").String())
			})
		}
	})

	t.Run("case=PATCH should not invalidate credentials ory/cloud#148", func(t *testing.T) {
		// see https://github.com/ory/cloud/issues/148

		createCredentials := func(t *testing.T) (*identity.Identity, string, string) {
			t.Helper()
			email := x.NewUUID().String() + "@ory.sh"
			password := "ljanf123akf"
			p, err := reg.Hasher(ctx).Generate(context.Background(), []byte(password))
			require.NoError(t, err)
			i := &identity.Identity{Traits: identity.Traits(`{"email":"` + email + `"}`)}
			i.SetCredentials(identity.CredentialsTypePassword, identity.Credentials{
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{email},
				Config:      sqlxx.JSONRawMessage(`{"hashed_password":"` + string(p) + `"}`),
			})
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
			return i, email, password
		}

		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				i, email, password := createCredentials(t)
				values := func(v url.Values) {
					v.Set("identifier", email)
					v.Set("password", password)
				}

				// verify login works initially
				loginResponse := testhelpers.SubmitLoginForm(t, true, ts.Client(), ts, values, false, true, 200, "")
				require.NotEmpty(t, gjson.Get(loginResponse, "session_token").String(), "expected to find a session token, found none")

				patch := []patch{
					{"op": "replace", "path": "/metadata_public", "value": map[string]string{"role": "user"}},
				}

				res := send(t, ts, "PATCH", "/identities/"+i.ID.String(), http.StatusOK, &patch)
				assert.EqualValues(t, "user", res.Get("metadata_public.role").String(), "%s", res.Raw)
				assert.NotEqualValues(t, i.StateChangedAt, sqlxx.NullTime(res.Get("state_changed_at").Time()), "%s", res.Raw)

				loginResponse = testhelpers.SubmitLoginForm(t, true, ts.Client(), ts, values, false, true, 200, "")
				msgs := gjson.Get(loginResponse, "ui.messages")
				require.Empty(t, msgs.Array(), "expected to find no messages: %s", msgs.String())
			})
		}
	})

	t.Run("case=PATCH should update metadata_admin correctly", func(t *testing.T) {
		i := &identity.Identity{Traits: identity.Traits(fmt.Sprintf(`{"subject":"%s"}`, x.NewUUID().String()))}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				patch := []patch{
					{"op": "add", "path": "/metadata_admin", "value": "metadata admin"},
				}

				res := send(t, ts, "PATCH", "/identities/"+i.ID.String(), http.StatusOK, &patch)

				assert.True(t, res.Get("metadata_admin").Exists(), "%s", res.Raw)
				assert.EqualValues(t, "metadata admin", res.Get("metadata_admin").String(), "%s", res.Raw)
			})
		}
	})

	t.Run("case=PATCH should update nested metadata_admin fields correctly", func(t *testing.T) {
		id := x.NewUUID().String()
		i := &identity.Identity{MetadataAdmin: sqlxx.NullJSONRawMessage(fmt.Sprintf(`{"id": "%s", "allowed": true}`, id))}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				patch := []patch{
					{"op": "replace", "path": "/metadata_admin/allowed", "value": "false"},
				}

				res := send(t, ts, "PATCH", "/identities/"+i.ID.String(), http.StatusOK, &patch)

				assert.True(t, res.Get("metadata_admin.allowed").Exists(), "%s", res.Raw)
				assert.EqualValues(t, false, res.Get("metadata_admin.allowed").Bool(), "%s", res.Raw)
				assert.EqualValues(t, id, res.Get("metadata_admin.id").String(), "%s", res.Raw)
			})
		}
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
				var cr identity.CreateIdentityBody
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
				var cr identity.CreateIdentityBody
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
				var cr identity.CreateIdentityBody
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
				var cr identity.CreateIdentityBody
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
				var cr identity.CreateIdentityBody
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
				var cr identity.CreateIdentityBody
				cr.SchemaID = "employee"
				originalEmail := x.NewUUID().String() + "@ory.sh"
				cr.Traits = []byte(`{"email":"` + originalEmail + `"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)
				assert.EqualValues(t, originalEmail, res.Get("recovery_addresses.0.value").String(), "%s", res.Raw)
				assert.EqualValues(t, originalEmail, res.Get("verifiable_addresses.0.value").String(), "%s", res.Raw)

				id := res.Get("id").String()
				updatedEmail := x.NewUUID().String() + "@ory.sh"
				res = send(t, ts, "PUT", "/identities/"+id, http.StatusOK, &identity.UpdateIdentityBody{
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
				var cr identity.CreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh", "department": "ory"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)

				id := res.Get("id").String()
				res = send(t, ts, "PUT", "/identities/"+id, http.StatusBadRequest, &identity.UpdateIdentityBody{
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
				var cr identity.CreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh", "department": "ory"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)

				id := res.Get("id").String()
				res = send(t, ts, "PUT", "/identities/"+id, http.StatusBadRequest, &identity.UpdateIdentityBody{
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
				var cr identity.CreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh", "department": "ory"}`)
				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)

				id := res.Get("id").String()
				res = send(t, ts, "PUT", "/identities/"+id, http.StatusOK, &identity.UpdateIdentityBody{
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
					var cr identity.CreateIdentityBody
					cr.SchemaID = "employee"
					cr.Traits = []byte(`{"department": "ory"}`)
					res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)

					id := res.Get("id").String()
					_ = send(t, ts, "PUT", "/identities/"+id, http.StatusOK, &identity.UpdateIdentityBody{
						SchemaID: "employee",
						Traits:   []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh"}`),
					})

					_ = send(t, ts, "PUT", "/identities/"+id, http.StatusOK, &identity.UpdateIdentityBody{
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
				var cr identity.CreateIdentityBody
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
		for name, ts := range map[string]*httptest.Server{"admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				res := get(t, ts, "/identities", http.StatusOK)
				assert.False(t, res.Get("0.credentials").Exists(), "credentials config should be omitted: %s", res.Raw)
				assert.True(t, res.Get("0.metadata_public").Exists(), "metadata_public config should be included: %s", res.Raw)
				assert.True(t, res.Get("0.metadata_admin").Exists(), "metadata_admin config should be included: %s", res.Raw)
				assert.EqualValues(t, "baz", res.Get(`#(traits.bar=="baz").traits.bar`).String(), "%s", res.Raw)
			})
		}
	})

	t.Run("organizations", func(t *testing.T) {
		t.Run("case=should list organization identities", func(t *testing.T) {
			for name, ts := range map[string]*httptest.Server{"admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					orgID := uuid.Must(uuid.NewV4())
					email := x.NewUUID().String() + "@ory.sh"
					require.NoError(t, reg.IdentityManager().Create(ctx, &identity.Identity{
						Traits:         identity.Traits(`{"email":"` + email + `"}`),
						OrganizationID: uuid.NullUUID{UUID: orgID, Valid: true},
					}))

					res := get(t, ts, "/identities?organization_id="+orgID.String(), http.StatusOK)
					assert.Len(t, res.Array(), 1)
					assert.EqualValues(t, email, res.Get(`0.traits.email`).String(), "%s", res.Raw)
				})
			}
		})

		t.Run("case=malformed organization id should return an error", func(t *testing.T) {
			for name, ts := range map[string]*httptest.Server{"admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					res := get(t, ts, "/identities?organization_id=not-a-uuid", http.StatusBadRequest)
					assert.Contains(t, res.Get("error.reason").String(), "Invalid UUID value `not-a-uuid` for parameter `organization_id`.", "%s", res.Raw)
				})
			}
		})

		t.Run("case=unknown organization id should return an empty list", func(t *testing.T) {
			for name, ts := range map[string]*httptest.Server{"admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					id := x.NewUUID()
					res := get(t, ts, "/identities?organization_id="+id.String(), http.StatusOK)
					assert.Len(t, res.Array(), 0)
				})
			}
		})
	})

	t.Run("case=should list all identities with credentials", func(t *testing.T) {
		t.Run("include_credential=oidc should include OIDC credentials config", func(t *testing.T) {
			res := get(t, adminTS, "/identities?include_credential=oidc&credentials_identifier=bar:foo.oidc@bar.com", http.StatusOK)
			assert.True(t, res.Get("0.credentials.oidc.config").Exists(), "credentials config should be included: %s", res.Raw)
			snapshotx.SnapshotT(t, res.Get("0.credentials.oidc.config").String())
		})
		t.Run("include_credential=totp should not include OIDC credentials config", func(t *testing.T) {
			res := get(t, adminTS, "/identities?include_credential=totp&credentials_identifier=bar:foo.oidc@bar.com", http.StatusOK)
			assert.False(t, res.Get("0.credentials.oidc.config").Exists(), "credentials config should be included: %s", res.Raw)
		})
	})

	t.Run("case=should not be able to list all identities with credentials due to wrong credentials type", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				res := get(t, ts, "/identities?include_credential=XYZ", http.StatusBadRequest)
				assert.Contains(t, res.Get("error.message").String(), "The request was malformed or contained invalid parameters", "%s", res.Raw)
			})
		}
	})

	t.Run("case=should list all identities with eventual consistency", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				res := get(t, ts, "/identities?consistency=eventual", http.StatusOK)
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

	t.Run("case=should not be able to patch an identity that does not exist yet", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				res := send(t, ts, "PATCH", "/identities/not-found", http.StatusNotFound, json.RawMessage(`{"traits": {"bar":"baz"}}`))
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

	t.Run("case=should delete credential of a specific user and no longer be able to retrieve it", func(t *testing.T) {
		ignoreDefault := []string{"id", "schema_url", "state_changed_at", "created_at", "updated_at"}
		type M = map[identity.CredentialsType]identity.Credentials
		createIdentity := func(creds M) func(*testing.T) *identity.Identity {
			return func(t *testing.T) *identity.Identity {
				i := identity.NewIdentity("")
				for k, v := range creds {
					v.Type = k
					creds[k] = v
				}
				i.Credentials = creds
				i.Traits = identity.Traits("{}")
				require.NoError(t, reg.Persister().CreateIdentity(context.Background(), i))
				return i
			}
		}
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("type=remove unknown identity/"+name, func(t *testing.T) {
				remove(t, ts, "/identities/"+x.NewUUID().String()+"/credentials/azerty", http.StatusNotFound)
			})
			t.Run("type=remove unknown type/"+name, func(t *testing.T) {
				i := createIdentity(M{
					identity.CredentialsTypePassword: {
						Config:      []byte(`{"hashed_password":"some_valid_hash"}`),
						Identifiers: []string{x.NewUUID().String()},
					},
				})(t)
				remove(t, ts, "/identities/"+i.ID.String()+"/credentials/azerty", http.StatusNotFound)
			})
			t.Run("type=deny to remove password type/"+name, func(t *testing.T) {
				i := createIdentity(M{
					identity.CredentialsTypePassword: {
						Config:      []byte(`{"hashed_password":"some_valid_hash"}`),
						Identifiers: []string{x.NewUUID().String()},
					},
				})(t)
				remove(t, ts, "/identities/"+i.ID.String()+"/credentials/password", http.StatusBadRequest)
			})
			t.Run("type=allow to remove password type/"+name, func(t *testing.T) {
				sub := x.NewUUID().String()
				pwIdentifier := x.NewUUID().String()
				i := createIdentity(M{
					identity.CredentialsTypePassword: {
						Config:      []byte(`{"hashed_password":"some_valid_hash"}`),
						Identifiers: []string{pwIdentifier},
					},
					identity.CredentialsTypeOIDC: {
						Config:      []byte(fmt.Sprintf(`{"providers":[{"subject":"%s","provider":"gh"}]}`, sub)),
						Identifiers: []string{identity.OIDCUniqueID("gh", sub)},
					},
				})(t)
				remove(t, ts, "/identities/"+i.ID.String()+"/credentials/password", http.StatusNoContent)
				actual, creds, err := reg.PrivilegedIdentityPool().FindByCredentialsIdentifier(ctx, identity.CredentialsTypePassword, pwIdentifier)
				require.NoError(t, err)
				assert.Equal(t, "{}", string(creds.Config))
				assert.Equal(t, i.ID, actual.ID)
			})
			t.Run("type=remove oidc type/"+name, func(t *testing.T) {
				// force ordering among github identifiers
				githubSubject := "0" + randx.MustString(7, randx.Numeric)
				githubSubject2 := "1" + randx.MustString(7, randx.Numeric)
				googleSubject := randx.MustString(8, randx.Numeric)
				initialConfig := []byte(fmt.Sprintf(`{
					"providers": [
						{
							"subject": %q,
							"provider": "github"
						},
						{
							"subject": %q,
							"provider": "github"
						},
						{
							"subject": %q,
							"provider": "google"
						}
					]
				}`, githubSubject, githubSubject2, googleSubject))
				identifiers := []string{
					identity.OIDCUniqueID("github", githubSubject),
					identity.OIDCUniqueID("github", githubSubject2),
					identity.OIDCUniqueID("google", googleSubject),
				}
				i := createIdentity(M{
					identity.CredentialsTypeOIDC: {
						Identifiers: identifiers,
						Config:      initialConfig,
					},
				})(t)
				res := get(t, ts, "/identities/"+i.ID.String()+"?include_credential=oidc", http.StatusOK)
				assert.EqualValues(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)
				assert.Len(t, res.Get("credentials.oidc.identifiers").Array(), 3, "%s", res.Raw)
				assert.EqualValues(t, res.Get("credentials.oidc.identifiers.0").String(), identifiers[0], "%s", res.Raw)
				assert.EqualValues(t, res.Get("credentials.oidc.identifiers.1").String(), identifiers[1], "%s", res.Raw)
				assert.EqualValues(t, res.Get("credentials.oidc.identifiers.2").String(), identifiers[2], "%s", res.Raw)

				oidConfig := gjson.Parse(res.Get("credentials.oidc.config").String())
				assert.Len(t, res.Get("credentials.oidc.identifiers").Array(), 3, "%s", res.Raw)
				assert.EqualValues(t, oidConfig.Get("providers.0.provider").String(), "github", "%s", res.Raw)
				assert.EqualValues(t, oidConfig.Get("providers.0.subject").String(), githubSubject, "%s", res.Raw)
				assert.EqualValues(t, oidConfig.Get("providers.1.provider").String(), "github", "%s", res.Raw)
				assert.EqualValues(t, oidConfig.Get("providers.1.subject").String(), githubSubject2, "%s", res.Raw)
				assert.EqualValues(t, oidConfig.Get("providers.2.provider").String(), "google", "%s", res.Raw)
				assert.EqualValues(t, oidConfig.Get("providers.2.subject").String(), googleSubject, "%s", res.Raw)

				remove(t, ts, "/identities/"+i.ID.String()+"/credentials/oidc?identifier="+identifiers[1], http.StatusNoContent)
				res = get(t, ts, "/identities/"+i.ID.String()+"?include_credential=oidc", http.StatusOK)

				assert.EqualValues(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)
				assert.Len(t, res.Get("credentials.oidc.identifiers").Array(), 2, "%s", res.Raw)
				assert.EqualValues(t, res.Get("credentials.oidc.identifiers.0").String(), identifiers[0], "%s", res.Raw)
				assert.EqualValues(t, res.Get("credentials.oidc.identifiers.1").String(), identifiers[2], "%s", res.Raw)

				oidConfig = gjson.Parse(res.Get("credentials.oidc.config").String())
				assert.Len(t, res.Get("credentials.oidc.identifiers").Array(), 2, "%s", res.Raw)
				assert.EqualValues(t, oidConfig.Get("providers.0.provider").String(), "github", "%s", res.Raw)
				assert.EqualValues(t, oidConfig.Get("providers.0.subject").String(), githubSubject, "%s", res.Raw)
				assert.EqualValues(t, oidConfig.Get("providers.1.provider").String(), "google", "%s", res.Raw)
				assert.EqualValues(t, oidConfig.Get("providers.1.subject").String(), googleSubject, "%s", res.Raw)
			})
			t.Run("type=remove webauthn passwordless type/"+name, func(t *testing.T) {
				expected := `{"credentials":[{"id":"THTndqZP5Mjvae1BFvJMaMfEMm7O7HE1ju+7PBaYA7Y=","added_at":"2022-12-16T14:11:55Z","public_key":"pQECAyYgASFYIMJLQhJxQRzhnKPTcPCUODOmxYDYo2obrm9bhp5lvSZ3IlggXjhZvJaPUqF9PXqZqTdWYPR7R+b2n/Wi+IxKKXsS4rU=","display_name":"test","authenticator":{"aaguid":"rc4AAjW8xgpkiwsl8fBVAw==","sign_count":0,"clone_warning":false},"is_passwordless":true,"attestation_type":"none"}],"user_handle":"Ef5JiMpMRwuzauWs/9J0gQ=="}`
				i := createIdentity(M{identity.CredentialsTypeWebAuthn: {Config: []byte(expected)}})(t)
				remove(t, ts, "/identities/"+i.ID.String()+"/credentials/webauthn", http.StatusNoContent)
				// Check that webauthn has not been deleted
				res := get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
				assert.EqualValues(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)

				actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
				require.NoError(t, err)
				snapshotx.SnapshotT(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), snapshotx.ExceptNestedKeys(append(ignoreDefault, "hashed_password")...), snapshotx.ExceptPaths("credentials.oidc.identifiers"))
			})
			t.Run("type=remove webauthn passwordless and multiple fido mfa type/"+name, func(t *testing.T) {
				message, err := json.Marshal(identity.CredentialsWebAuthnConfig{
					Credentials: identity.CredentialsWebAuthn{
						{
							// Passwordless 1
							ID:          []byte("THTndqZP5Mjvae1BFvJMaMfEMm7O7HE1ju+7PBaYA7Y="),
							AddedAt:     time.Date(2022, 12, 16, 14, 11, 55, 0, time.UTC),
							PublicKey:   []byte("pQECAyYgASFYIMJLQhJxQRzhnKPTcPCUODOmxYDYo2obrm9bhp5lvSZ3IlggXjhZvJaPUqF9PXqZqTdWYPR7R+b2n/Wi+IxKKXsS4rU="),
							DisplayName: "test",
							Authenticator: &identity.AuthenticatorWebAuthn{
								AAGUID:       []byte("rc4AAjW8xgpkiwsl8fBVAw=="),
								SignCount:    0,
								CloneWarning: false,
							},
							IsPasswordless:  true,
							AttestationType: "none",
						}, {
							// Passwordless 2
							ID:          []byte("THTndqZP5Mjvae1BFvJMaMfEMm7O7HE2ju+7PBaYA7Y="),
							AddedAt:     time.Date(2022, 12, 16, 14, 11, 55, 0, time.UTC),
							PublicKey:   []byte("pQECAyYgASFYIMJLQhJxQRzhnKPTcPCUODOmxYDYo2obrm9bhp5lvSZ3IlggXjhZvJaPUqF9PXqZqTdWYPR7R+b2n/Wi+IxKKXsS4rU="),
							DisplayName: "test",
							Authenticator: &identity.AuthenticatorWebAuthn{
								AAGUID:       []byte("rc4AAjW8xgpkiwsl8fBVAw=="),
								SignCount:    0,
								CloneWarning: false,
							},
							IsPasswordless:  true,
							AttestationType: "none",
						}, {
							// MFA 1
							ID:          []byte("THTndqZP5Mjvae1BFvJMaMfEMm7O7HE3ju+7PBaYA7Y="),
							AddedAt:     time.Date(2022, 12, 16, 14, 11, 55, 0, time.UTC),
							PublicKey:   []byte("pQECAyYgASFYIMJLQhJxQRzhnKPTcPCUODOmxYDYo2obrm9bhp5lvSZ3IlggXjhZvJaPUqF9PXqZqTdWYPR7R+b2n/Wi+IxKKXsS4rU="),
							DisplayName: "test",
							Authenticator: &identity.AuthenticatorWebAuthn{
								AAGUID:       []byte("rc4AAjW8xgpkiwsl8fBVAw=="),
								SignCount:    0,
								CloneWarning: false,
							},
							IsPasswordless:  false,
							AttestationType: "none",
						}, {
							// MFA 2
							ID:          []byte("THTndqZP5Mjvae1BFvJMaMfEMm7O7HE4ju+7PBaYA7Y="),
							AddedAt:     time.Date(2022, 12, 16, 14, 11, 55, 0, time.UTC),
							PublicKey:   []byte("pQECAyYgASFYIMJLQhJxQRzhnKPTcPCUODOmxYDYo2obrm9bhp5lvSZ3IlggXjhZvJaPUqF9PXqZqTdWYPR7R+b2n/Wi+IxKKXsS4rU="),
							DisplayName: "test",
							Authenticator: &identity.AuthenticatorWebAuthn{
								AAGUID:       []byte("rc4AAjW8xgpkiwsl8fBVAw=="),
								SignCount:    0,
								CloneWarning: false,
							},
							IsPasswordless:  false,
							AttestationType: "none",
						},
					},
					UserHandle: []byte("Ef5JiMpMRwuzauWs/9J0gQ=="),
				})
				require.NoError(t, err)

				i := createIdentity(M{identity.CredentialsTypeWebAuthn: {Config: message}})(t)
				remove(t, ts, "/identities/"+i.ID.String()+"/credentials/webauthn", http.StatusNoContent)
				// Check that webauthn has not been deleted
				res := get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
				assert.EqualValues(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)

				actual, err := reg.PrivilegedIdentityPool().GetIdentityConfidential(ctx, uuid.FromStringOrNil(res.Get("id").String()))
				require.NoError(t, err)
				snapshotx.SnapshotT(t, identity.WithCredentialsAndAdminMetadataInJSON(*actual), snapshotx.ExceptNestedKeys(append(ignoreDefault, "hashed_password")...), snapshotx.ExceptPaths("credentials.oidc.identifiers"))
			})
			for ct, ctConf := range map[identity.CredentialsType][]byte{
				identity.CredentialsTypeLookup:   []byte(`{"recovery_codes": [{"code": "aaa"}]}`),
				identity.CredentialsTypeTOTP:     []byte(`{"totp_url":"otpauth://totp/test"}`),
				identity.CredentialsTypeWebAuthn: []byte(`{"credentials":[{"id":"THTndqZP5Mjvae1BFvJMaMfEMm7O7HE1ju+7PBaYA7Y=","added_at":"2022-12-16T14:11:55Z","public_key":"pQECAyYgASFYIMJLQhJxQRzhnKPTcPCUODOmxYDYo2obrm9bhp5lvSZ3IlggXjhZvJaPUqF9PXqZqTdWYPR7R+b2n/Wi+IxKKXsS4rU=","display_name":"test","authenticator":{"aaguid":"rc4AAjW8xgpkiwsl8fBVAw==","sign_count":0,"clone_warning":false},"is_passwordless":false,"attestation_type":"none"}],"user_handle":"Ef5JiMpMRwuzauWs/9J0gQ=="}`),
			} {
				t.Run("type=remove "+string(ct)+"/"+name, func(t *testing.T) {
					for _, tc := range []struct {
						desc  string
						exist bool
						setup func(t *testing.T) *identity.Identity
					}{
						{
							desc:  "with",
							exist: true,
							setup: createIdentity(M{
								identity.CredentialsTypePassword: {Config: []byte(`{"secret":"pst"}`)},
								ct:                               {Config: ctConf},
							}),
						},
						{
							desc:  "without",
							exist: false,
							setup: createIdentity(M{
								identity.CredentialsTypePassword: {Config: []byte(`{"secret":"pst"}`)},
							}),
						},
						{
							desc:  "multiple",
							exist: true,
							setup: createIdentity(M{
								identity.CredentialsTypePassword: {Config: []byte(`{"secret":"pst"}`)},
								identity.CredentialsTypeOIDC:     {Config: []byte(`{"id":"pst"}`)},
								ct:                               {Config: ctConf},
							}),
						},
					} {
						t.Run("type=remove "+string(ct)+"/"+name+"/"+tc.desc, func(t *testing.T) {
							i := tc.setup(t)
							credName := string(ct)
							// Initial Querying
							resBefore := get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
							assert.EqualValues(t, i.ID.String(), resBefore.Get("id").String(), "%s", resBefore.Raw)
							assert.True(t, resBefore.Get("credentials").Exists())
							if tc.exist {
								assert.True(t, resBefore.Get("credentials").Get(credName).Exists())
								// Remove
								remove(t, ts, "/identities/"+i.ID.String()+"/credentials/"+credName, http.StatusNoContent)
								// Query back
								resAfter := get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
								assert.EqualValues(t, i.ID.String(), resAfter.Get("id").String(), "%s", resAfter.Raw)
								assert.True(t, resAfter.Get("credentials").Exists())
								// Check results
								expected := resBefore.Get("credentials").Map()
								delete(expected, credName)
								expectedKeys := x.Keys(expected)
								sort.Strings(expectedKeys)
								result := resAfter.Get("credentials").Map()
								resultKeys := x.Keys(result)
								sort.Strings(resultKeys)
								assert.Equal(t, resultKeys, expectedKeys)
							} else {
								assert.False(t, resBefore.Get("credentials").Get(credName).Exists())
								remove(t, ts, "/identities/"+i.ID.String()+"/credentials/"+credName, http.StatusNotFound)
							}
						})
					}
				})
			}
		}
	})

	t.Run("case=should paginate all identities", func(t *testing.T) {
		// Start new server
		conf, reg := internal.NewFastRegistryWithMocks(t)
		_, ts := testhelpers.NewKratosServerWithCSRF(t, reg)
		mockServerURL := urlx.ParseOrPanic(publicTS.URL)
		conf.MustSet(ctx, config.ViperKeyAdminBaseURL, ts.URL)
		testhelpers.SetIdentitySchemas(t, conf, map[string]string{
			"default":         "file://./stub/identity.schema.json",
			"customer":        "file://./stub/handler/customer.schema.json",
			"multiple_emails": "file://./stub/handler/multiple_emails.schema.json",
			"employee":        "file://./stub/handler/employee.schema.json",
		})
		conf.MustSet(ctx, config.ViperKeyPublicBaseURL, mockServerURL.String())

		var toCreate []*identity.Identity
		count := 500
		for range count {
			i := identity.NewIdentity(config.DefaultIdentityTraitsSchemaID)
			i.Traits = identity.Traits(`{"email":"` + x.NewUUID().String() + `@ory.sh"}`)
			toCreate = append(toCreate, i)
		}

		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentities(context.Background(), toCreate...))

		for _, perPage := range []int{10, 50, 100, 500} {
			t.Run(fmt.Sprintf("perPage=%d", perPage), func(t *testing.T) {
				t.Parallel()
				body, _ := getFull(t, ts, fmt.Sprintf("/identities?per_page=%d", perPage), http.StatusOK)
				assert.Len(t, body.Array(), perPage)
			})
		}

		t.Run("iterate over next page", func(t *testing.T) {
			perPage := 10

			run := func(t *testing.T, path string, knownIDs map[string]struct{}) (next *url.URL, res *http.Response) {
				t.Logf("Requesting %s", path)
				body, res := getFull(t, ts, path, http.StatusOK)
				for _, i := range body.Array() {
					id := i.Get("id").String()
					_, seen := knownIDs[id]
					require.Falsef(t, seen, "ID %s was previously returned from the API", id)
					knownIDs[id] = struct{}{}
				}
				links := link.ParseResponse(res)
				if nextLink, ok := links["next"]; ok {
					next, err := url.Parse(nextLink.URI)
					require.NoError(t, err)
					return next, res
				}
				return nil, res
			}

			t.Run("using token pagination", func(t *testing.T) {
				knownIDs := make(map[string]struct{})
				var pages int
				path := fmt.Sprintf("/admin/identities?page_size=%d", perPage)
				for {
					pages++
					next, res := run(t, path, knownIDs)
					assert.NotContains(t, res.Header, "X-Total-Count", "not supported in token pagination")
					if next == nil {
						break
					}
					assert.NotContains(t, next.Query(), "page")
					assert.NotContains(t, next.Query(), "per_page")
					path = next.Path + "?" + next.Query().Encode()
				}

				assert.Len(t, knownIDs, count)
				assert.Equal(t, count/perPage, pages)
			})

			t.Run("using page pagination", func(t *testing.T) {
				knownIDs := make(map[string]struct{})
				var pages int
				path := fmt.Sprintf("/admin/identities?page=0&per_page=%d", perPage)
				for {
					pages++
					next, res := run(t, path, knownIDs)
					assert.Equal(t, strconv.Itoa(count), res.Header.Get("X-Total-Count"))
					if next == nil {
						break
					}
					path = next.Path + "?" + next.Query().Encode()
				}

				assert.Len(t, knownIDs, count)
				assert.Equal(t, count/perPage, pages)
			})
		})
	})
}

func validCreateIdentityBody(prefix string, i int) *identity.CreateIdentityBody {
	var (
		verifiableAddresses []identity.VerifiableAddress
		recoveryAddresses   []identity.RecoveryAddress
	)
	traits := struct {
		Emails   []string `json:"emails"`
		Username string   `json:"username"`
	}{}

	verificationStates := []identity.VerifiableAddressStatus{
		identity.VerifiableAddressStatusPending,
		identity.VerifiableAddressStatusSent,
		identity.VerifiableAddressStatusCompleted,
	}

	for j := range 4 {
		email := fmt.Sprintf("%s-%d-%d@ory.sh", prefix, i, j)
		traits.Emails = append(traits.Emails, email)
		verifiableAddresses = append(verifiableAddresses, identity.VerifiableAddress{
			Value:    email,
			Via:      identity.VerifiableAddressTypeEmail,
			Verified: j%2 == 0,
			Status:   verificationStates[j%len(verificationStates)],
		})
		recoveryAddresses = append(recoveryAddresses, identity.RecoveryAddress{
			Value: email,
			Via:   identity.RecoveryAddressTypeEmail,
		})
	}
	traits.Username = traits.Emails[0]
	rawTraits, _ := json.Marshal(traits)

	return &identity.CreateIdentityBody{
		SchemaID: "multiple_emails",
		Traits:   rawTraits,
		Credentials: &identity.IdentityWithCredentials{
			Password: &identity.AdminIdentityImportCredentialsPassword{
				Config: identity.AdminIdentityImportCredentialsPasswordConfig{
					Password: fmt.Sprintf("password-%d", i),
				},
			},
		},
		VerifiableAddresses: verifiableAddresses,
		RecoveryAddresses:   recoveryAddresses,
		MetadataPublic:      json.RawMessage(fmt.Sprintf(`{"public-%d":"public"}`, i)),
		MetadataAdmin:       json.RawMessage(fmt.Sprintf(`{"admin-%d":"admin"}`, i)),
		State:               "active",
	}
}

func assertJSONArrayElementsMatch(t *testing.T, expected, actual gjson.Result, msgAndArgs ...any) {
	t.Helper()

	var expectedStrings, actualStrings []string
	expected.ForEach(func(_, value gjson.Result) bool {
		expectedStrings = append(expectedStrings, value.String())
		return true
	})
	actual.ForEach(func(_, value gjson.Result) bool {
		actualStrings = append(actualStrings, value.String())
		return true
	})

	assert.ElementsMatch(t, expectedStrings, actualStrings, msgAndArgs...)
}
