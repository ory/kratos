package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/faker/v3"

	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
)

func TestHandler(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)

	// Start kratos server
	publicTS, adminTS := testhelpers.NewKratosServerWithCSRF(t, reg)

	mockServerURL := urlx.ParseOrPanic(publicTS.URL)
	defaultSchemaExternalURL := (&schema.Schema{ID: "default"}).SchemaURL(mockServerURL).String()

	conf.MustSet(config.ViperKeyAdminBaseURL, adminTS.URL)
	testhelpers.SetDefaultIdentitySchema(t, conf, "file://./stub/identity.schema.json")
	testhelpers.SetIdentitySchemas(t, conf, map[string]string{
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
		require.NoError(t, json.NewEncoder(&b).Encode(send))
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
		t.Run("case=should create an identity with an ID which is ignored", func(t *testing.T) {
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					res := send(t, ts, "POST", "/identities", http.StatusCreated, json.RawMessage(`{"traits": {"bar":"baz"}}`))

					i.Traits = []byte(res.Get("traits").Raw)
					i.ID = x.ParseUUID(res.Get("id").String())
					i.StateChangedAt = sqlxx.NullTime(res.Get("state_changed_at").Time())
					assert.NotEmpty(t, res.Get("id").String())

					assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
					assert.Empty(t, res.Get("credentials").String(), "%s", res.Raw)
					assert.EqualValues(t, defaultSchemaExternalURL, res.Get("schema_url").String(), "%s", res.Raw)
					assert.EqualValues(t, config.DefaultIdentityTraitsSchemaID, res.Get("schema_id").String(), "%s", res.Raw)
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
					assert.Empty(t, res.Get("credentials").String(), "%s", res.Raw)
				})
			}
		})

		t.Run("case=should update an identity and persist the changes", func(t *testing.T) {
			for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
				t.Run("endpoint="+name, func(t *testing.T) {
					ur := identity.AdminUpdateIdentityBody{
						Traits:   []byte(`{"bar":"baz","foo":"baz"}`),
						SchemaID: i.SchemaID,
						State:    identity.StateInactive,
					}

					res := send(t, ts, "PUT", "/identities/"+i.ID.String(), http.StatusOK, &ur)
					assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
					assert.EqualValues(t, "baz", res.Get("traits.foo").String(), "%s", res.Raw)
					assert.EqualValues(t, identity.StateInactive, res.Get("state").String(), "%s", res.Raw)
					assert.NotEqualValues(t, i.StateChangedAt, sqlxx.NullTime(res.Get("state_changed_at").Time()), "%s", res.Raw)

					res = get(t, ts, "/identities/"+i.ID.String(), http.StatusOK)
					assert.EqualValues(t, i.ID.String(), res.Get("id").String(), "%s", res.Raw)
					assert.EqualValues(t, "baz", res.Get("traits.bar").String(), "%s", res.Raw)
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

	t.Run("case=should create an identity with a different schema", func(t *testing.T) {
		for name, ts := range map[string]*httptest.Server{"public": publicTS, "admin": adminTS} {
			t.Run("endpoint="+name, func(t *testing.T) {
				var cr identity.AdminCreateIdentityBody
				cr.SchemaID = "employee"
				cr.Traits = []byte(`{"email":"` + x.NewUUID().String() + `@ory.sh"}`)

				res := send(t, ts, "POST", "/identities", http.StatusCreated, &cr)
				assert.JSONEq(t, string(cr.Traits), res.Get("traits").Raw, "%s", res.Raw)
				assert.EqualValues(t, "employee", res.Get("schema_id").String(), "%s", res.Raw)
				assert.EqualValues(t, mockServerURL.String()+"/schemas/employee", res.Get("schema_url").String(), "%s", res.Raw)
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
				assert.EqualValues(t, mockServerURL.String()+"/schemas/employee", res.Get("schema_url").String(), "%s", res.Raw)
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
