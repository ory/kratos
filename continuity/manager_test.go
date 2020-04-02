package continuity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"
	"github.com/ory/viper"
	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/continuity"
	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/x"
)

type persisterTestCase struct {
	ro          []continuity.ManagerOption
	wo          []continuity.ManagerOption
	expected    interface{}
	expectedErr error
}

type persisterTestPayload struct {
	Foo string `json:"foo"`
}

func TestManager(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)

	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/identity.schema.json")
	viper.Set(configuration.ViperKeyURLsSelfPublic, "https://www.ory.sh")
	i := identity.NewIdentity("")
	require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))

	var newServer = func(t *testing.T, p continuity.Manager, tc *persisterTestCase) *httptest.Server {
		writer := herodot.NewJSONWriter(logrusx.New())
		router := httprouter.New()
		router.PUT("/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			if err := p.Pause(r.Context(), w, r, ps.ByName("name"), tc.ro...); err != nil {
				writer.WriteError(w, r, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		router.POST("/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			if err := p.Pause(r.Context(), w, r, ps.ByName("name"), tc.ro...); err != nil {
				writer.WriteError(w, r, err)
				return
			}

			c, err := p.Continue(r.Context(), r, ps.ByName("name"), tc.wo...)
			if err != nil {
				writer.WriteError(w, r, err)
				return
			}
			writer.Write(w, r, c)
		})

		router.GET("/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			c, err := p.Continue(r.Context(), r, ps.ByName("name"), tc.ro...)
			if err != nil {
				writer.WriteError(w, r, err)
				return
			}
			writer.Write(w, r, c)
		})

		ts := httptest.NewServer(router)
		t.Cleanup(func() {
			ts.Close()
		})
		return ts
	}

	var newClient = func() *http.Client {
		return &http.Client{Jar: x.EasyCookieJar(t, nil)}
	}

	for name, p := range map[string]continuity.Manager{
		"cookie": reg.ContinuityManager(),
	} {
		t.Run(fmt.Sprintf("persister=%s", name), func(t *testing.T) {
			for k, tc := range []persisterTestCase{
				{},
				{
					ro:       []continuity.ManagerOption{continuity.WithPayload(&persisterTestPayload{"bar"})},
					wo:       []continuity.ManagerOption{continuity.WithPayload(&persisterTestPayload{})},
					expected: &persisterTestPayload{"bar"},
				},
				{
					ro: []continuity.ManagerOption{continuity.WithIdentity(i)},
					wo: []continuity.ManagerOption{continuity.WithIdentity(i)},
				},
			} {
				t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
					cl := newClient()
					ts := newServer(t, p, &tc)
					var genid = func() string {
						return ts.URL + "/" + x.NewUUID().String()
					}

					t.Run("case=resume non-existing session", func(t *testing.T) {
						href := genid()
						res, err := cl.Do(x.NewTestHTTPRequest(t, "GET", href, nil))
						require.NoError(t, err)
						t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

						body := x.MustReadAll(res.Body)
						require.Equal(t, http.StatusBadRequest, res.StatusCode)
						assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), "resumable session")
					})

					t.Run("case=pause and resume session", func(t *testing.T) {
						href := genid()
						res, err := cl.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
						require.NoError(t, err)
						require.NoError(t, res.Body.Close())
						require.Equal(t, http.StatusNoContent, res.StatusCode)

						res, err = cl.Do(x.NewTestHTTPRequest(t, "GET", href, nil))
						require.NoError(t, err)
						t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

						body := x.MustReadAll(res.Body)
						if tc.expectedErr != nil {
							require.Equal(t, http.StatusGone, res.StatusCode, "%s", body)
							return
						}

						require.Equal(t, http.StatusOK, res.StatusCode, "%s", body)

						var b bytes.Buffer
						require.NoError(t, json.NewEncoder(&b).Encode(tc.expected))
						assert.JSONEq(t, b.String(), gjson.GetBytes(body, "payload").Raw, "%s", body)
						assert.Contains(t, href, gjson.GetBytes(body, "name").String(), "%s", body)
					})

					t.Run("case=pause and abort session", func(t *testing.T) {
						href := genid()
						res, err := cl.Do(x.NewTestHTTPRequest(t, "PUT", href, nil))
						require.NoError(t, err)
						require.NoError(t, res.Body.Close())
						require.Equal(t, http.StatusNoContent, res.StatusCode)

						res, err = cl.Do(x.NewTestHTTPRequest(t, "GET", href, nil))
						require.NoError(t, err)
						t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

						res, err = cl.Do(x.NewTestHTTPRequest(t, "GET", href, nil))
						require.NoError(t, err)
						require.Equal(t, http.StatusBadRequest, res.StatusCode)
						body := x.MustReadAll(res.Body)
						t.Cleanup(func() { require.NoError(t, res.Body.Close()) })
						assert.Contains(t, gjson.GetBytes(body, "error.reason").String(), "resumable session")
					})

					t.Run("case=pause and resume session in the same request", func(t *testing.T) {
						href := genid()
						res, err := cl.Do(x.NewTestHTTPRequest(t, "POST", href, nil))
						require.NoError(t, err)
						require.Equal(t, http.StatusOK, res.StatusCode)
						t.Cleanup(func() { require.NoError(t, res.Body.Close()) })

						var b bytes.Buffer
						require.NoError(t, json.NewEncoder(&b).Encode(tc.expected))

						body := x.MustReadAll(res.Body)
						assert.JSONEq(t, b.String(), gjson.GetBytes(body, "payload").Raw, "%s", body)
						assert.Contains(t, href, gjson.GetBytes(body, "name").String(), "%s", body)
					})
				})
			}
		})
	}
}
