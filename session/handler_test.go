// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session_test

import (
	"context"
	"encoding/base64"
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
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/corpx"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	. "github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/kratos/x/nosurfx"
	"github.com/ory/x/configx"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/pagination/keysetpagination"
	"github.com/ory/x/sqlcon"
	"github.com/ory/x/sqlxx"
	"github.com/ory/x/urlx"
)

func init() {
	corpx.RegisterFakes()
}

func send(code int) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(code)
	}
}

func TestSessionWhoAmI(t *testing.T) {
	t.Parallel()

	conf, reg := internal.NewFastRegistryWithMocks(t)
	ts, _, r, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)
	ctx := context.Background()

	// set this intermediate because kratos needs some valid url for CRUDE operations
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "http://example.com")
	email := "foo" + uuid.Must(uuid.NewV4()).String() + "@bar.sh"
	externalID := x.NewUUID().String()
	i := &identity.Identity{
		ID:         x.NewUUID(),
		ExternalID: sqlxx.NullString(externalID),
		State:      identity.StateActive,
		Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {
				Type:        identity.CredentialsTypePassword,
				Identifiers: []string{x.NewUUID().String()},
				Config:      []byte(`{"hashed_password":"$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`),
			},
		},
		Traits:         identity.Traits(`{"email": "` + email + `","baz":"bar","foo":true,"bar":2.5}`),
		MetadataAdmin:  []byte(`{"admin":"ma"}`),
		MetadataPublic: []byte(`{"public":"mp"}`),
		RecoveryAddresses: []identity.RecoveryAddress{
			{
				Value: email,
				Via:   identity.AddressTypeEmail,
			},
		},
		VerifiableAddresses: []identity.VerifiableAddress{
			{
				Value: email,
				Via:   identity.AddressTypeEmail,
			},
		},
	}
	h, _ := testhelpers.MockSessionCreateHandlerWithIdentity(t, reg, i)

	r.GET("/set", h)
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, ts.URL)

	t.Run("case=aal requirements", func(t *testing.T) {
		h1, _ := testhelpers.MockSessionCreateHandlerWithIdentityAndAMR(t, reg,
			newAAL2Identity(),
			[]identity.CredentialsType{identity.CredentialsTypePassword, identity.CredentialsTypeWebAuthn})
		r.GET("/set/aal2-aal2", h1)

		h2, _ := testhelpers.MockSessionCreateHandlerWithIdentityAndAMR(t, reg,
			newAAL2Identity(),
			[]identity.CredentialsType{identity.CredentialsTypePassword})
		r.GET("/set/aal2-aal1", h2)

		h3, _ := testhelpers.MockSessionCreateHandlerWithIdentityAndAMR(t, reg,
			newAAL1Identity(),
			[]identity.CredentialsType{identity.CredentialsTypePassword})
		r.GET("/set/aal1-aal1", h3)

		run := func(t *testing.T, endpoint string, kind string, code int) string {
			client := testhelpers.NewClientWithCookies(t)
			testhelpers.MockHydrateCookieClient(t, client, ts.URL+"/set/"+kind)

			res, err := client.Get(ts.URL + endpoint)
			require.NoError(t, err)
			body := x.MustReadAll(res.Body)
			assert.EqualValues(t, code, res.StatusCode)
			return string(body)
		}

		for k, e := range map[string]string{
			"whoami":     RouteWhoami,
			"collection": RouteCollection,
		} {
			t.Run(fmt.Sprintf("endpoint=%s", k), func(t *testing.T) {
				t.Run("case=aal2-aal2", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
					run(t, e, "aal2-aal2", http.StatusOK)
				})

				t.Run("case=aal2-aal2", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, "aal1")
					run(t, e, "aal2-aal2", http.StatusOK)
				})

				t.Run("case=aal2-aal1", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
					body := run(t, e, "aal2-aal1", http.StatusForbidden)
					assert.EqualValues(t, NewErrAALNotSatisfied("").Reason(), gjson.Get(body, "error.reason").String(), body)
				})

				t.Run("case=aal2-aal1", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, "aal1")
					run(t, e, "aal2-aal1", http.StatusOK)
				})

				t.Run("case=aal1-aal1", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
					run(t, e, "aal1-aal1", http.StatusOK)
				})
			})
		}
	})

	t.Run("case=http methods", func(t *testing.T) {
		run := func(t *testing.T, cacheEnabled bool, maxAge time.Duration) {
			conf.MustSet(ctx, config.ViperKeySessionWhoAmICaching, cacheEnabled)
			conf.MustSet(ctx, config.ViperKeySessionWhoAmICachingMaxAge, maxAge)
			client := testhelpers.NewClientWithCookies(t)

			// No cookie yet -> 401
			res, err := client.Get(ts.URL + RouteWhoami)
			require.NoError(t, err)
			testhelpers.AssertNoCSRFCookieInResponse(t, ts, client, res) // Test that no CSRF cookie is ever set here.

			if cacheEnabled {
				assert.NotEmpty(t, res.Header.Get("Ory-Session-Cache-For"))
				assert.Equal(t, "60", res.Header.Get("Ory-Session-Cache-For"))
			} else {
				assert.Empty(t, res.Header.Get("Ory-Session-Cache-For"))
			}

			// Set cookie
			reg.CSRFHandler().IgnorePath("/set")
			testhelpers.MockHydrateCookieClient(t, client, ts.URL+"/set")

			// Cookie set -> 200 (GET)
			for _, method := range []string{
				"GET",
				"POST",
				"PUT",
				"DELETE",
			} {
				t.Run("http_method="+method, func(t *testing.T) {
					req, err := http.NewRequest(method, ts.URL+RouteWhoami, nil)
					require.NoError(t, err)

					res, err = client.Do(req)
					require.NoError(t, err)
					body, err := io.ReadAll(res.Body)
					require.NoError(t, err)
					testhelpers.AssertNoCSRFCookieInResponse(t, ts, client, res) // Test that no CSRF cookie is ever set here.

					assert.EqualValues(t, http.StatusOK, res.StatusCode)
					assert.NotEmpty(t, res.Header.Get("X-Kratos-Authenticated-Identity-Id"))

					if cacheEnabled {
						var expectedSeconds int
						if maxAge > 0 {
							expectedSeconds = int(maxAge.Seconds())
						} else {
							expectedSeconds = int(conf.SessionLifespan(ctx).Seconds())
						}
						assert.InDelta(t, expectedSeconds, x.Must(strconv.Atoi(res.Header.Get("Ory-Session-Cache-For"))), 5)
					} else {
						assert.Empty(t, res.Header.Get("Ory-Session-Cache-For"))
					}

					assert.Empty(t, gjson.GetBytes(body, "identity.credentials"))
					assert.Equal(t, "mp", gjson.GetBytes(body, "identity.metadata_public.public").String(), "%s", body)
					assert.False(t, gjson.GetBytes(body, "identity.metadata_admin").Exists())

					assert.NotEmpty(t, gjson.GetBytes(body, "identity.recovery_addresses").String(), "%s", body)
					assert.NotEmpty(t, gjson.GetBytes(body, "identity.verifiable_addresses").String(), "%s", body)

					assert.Equal(t, externalID, gjson.GetBytes(body, "identity.external_id").String(), "%s", body)
				})
			}
		}

		t.Run("cache disabled", func(t *testing.T) {
			run(t, false, 0)
		})

		t.Run("cache enabled", func(t *testing.T) {
			run(t, true, 0)
		})

		t.Run("cache enabled with max age", func(t *testing.T) {
			run(t, true, time.Minute)
		})
	})

	t.Run("tokenize", func(t *testing.T) {
		setTokenizeConfig(conf, "es256", "jwk.es256.json", "")
		conf.MustSet(ctx, config.ViperKeySessionWhoAmICaching, true)

		h3, _ := testhelpers.MockSessionCreateHandlerWithIdentityAndAMR(t, reg, newAAL1Identity(), []identity.CredentialsType{identity.CredentialsTypePassword})
		r.GET("/set/tokenize", h3)

		client := testhelpers.NewClientWithCookies(t)
		testhelpers.MockHydrateCookieClient(t, client, ts.URL+"/set/"+"tokenize")

		res, err := client.Get(ts.URL + RouteWhoami + "?tokenize_as=es256")
		require.NoError(t, err)
		body := x.MustReadAll(res.Body)
		assert.EqualValues(t, http.StatusOK, res.StatusCode, string(body))

		token := gjson.GetBytes(body, "tokenized").String()
		require.NotEmpty(t, token)
		segments := strings.Split(token, ".")
		require.Len(t, segments, 3, token)
		decoded, err := base64.RawURLEncoding.DecodeString(segments[1])
		require.NoError(t, err)

		assert.NotEmpty(t, gjson.GetBytes(decoded, "sub").Str, decoded)
		assert.Empty(t, res.Header.Get("Ory-Session-Cache-For"))
	})

	/*
		t.Run("case=respects AAL config", func(t *testing.T) {
			conf.MustSet(ctx, config.ViperKeySessionLifespan, "1m")

			t.Run("required_aal=aal1", func(t *testing.T) {
				conf.MustSet(ctx, config.ViperKeySelfServiceSettingsRequiredAAL, "aal1")

				i := identity.Identity{Traits: []byte("{}"), State: identity.StateActive}
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
				s, err := testhelpers.NewActiveSession(&i, conf, time.Now(), identity.CredentialsTypePassword)
				require.NoError(t, err)
				require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), s))
				require.NotEmpty(t, s.Token)

				req, err := http.NewRequest("GET", pts.URL+"/session/get", nil)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+s.Token)

				c := http.DefaultClient
				res, err := c.Do(req)
				require.NoError(t, err)
				assert.EqualValues(t, http.StatusOK, res.StatusCode)
			})

			t.Run("required_aal=aal2", func(t *testing.T) {
				idAAL2 := identity.Identity{Traits: []byte("{}"), State: identity.StateActive, Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Config: []byte("{}")},
					identity.CredentialsTypeWebAuthn: {Type: identity.CredentialsTypeWebAuthn, Config: []byte("{}")},
				}}
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &idAAL2))

				idAAL1 := identity.Identity{Traits: []byte("{}"), State: identity.StateActive, Credentials: map[identity.CredentialsType]identity.Credentials{
					identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Config: []byte("{}")},
				}}
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &idAAL1))

				run := func(t *testing.T, complete []identity.CredentialsType, expectedCode int, i *identity.Identity) {

					s := session.NewInactiveSession()
					for _, m := range complete {
						s.CompletedLoginFor(m)
					}
					require.NoError(t, s.Activate(i, conf, time.Now().UTC()))

					require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), s))
					require.NotEmpty(t, s.Token)

					req, err := http.NewRequest("GET", pts.URL+"/session/get", nil)
					require.NoError(t, err)
					req.Header.Set("Authorization", "Bearer "+s.Token)

					c := http.DefaultClient
					res, err := c.Do(req)
					require.NoError(t, err)
					assert.EqualValues(t, expectedCode, res.StatusCode)
				}

				t.Run("fulfilled for aal2 if identity has aal2", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword, identity.CredentialsTypeWebAuthn}, 200, &idAAL2)
				})

				t.Run("rejected for aal1 if identity has aal2", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword}, 403, &idAAL2)
				})

				t.Run("fulfilled for aal1 if identity has aal2 but config is aal1", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, "aal1")
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword}, 200, &idAAL2)
				})

				t.Run("fulfilled for aal2 if identity has aal1", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword, identity.CredentialsTypeWebAuthn}, 200, &idAAL1)
				})

				t.Run("fulfilled for aal1 if identity has aal1", func(t *testing.T) {
					conf.MustSet(ctx, config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword}, 200, &idAAL1)
				})
			})
		})
	*/
}

func TestIsNotAuthenticatedSecurecookie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	r := x.NewRouterPublic(reg)
	r.GET("/public/with-callback", reg.SessionHandler().IsNotAuthenticated(send(http.StatusOK), send(http.StatusBadRequest)))

	ts := httptest.NewServer(r)
	defer ts.Close()
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, ts.URL)

	c := testhelpers.NewClientWithCookies(t)
	c.Jar.SetCookies(urlx.ParseOrPanic(ts.URL), []*http.Cookie{
		{
			Name: config.DefaultSessionCookieName,
			// This is an invalid cookie because it is generated by a very random secret
			Value:    "MTU3Mjg4Njg0MXxEdi1CQkFFQ180SUFBUkFCRUFBQU52LUNBQUVHYzNSeWFXNW5EQVVBQTNOcFpBWnpkSEpwYm1jTUd3QVpUWFZXVUhSQlZVeExXRWRUUmxkVVoyUkpUVXhzY201SFNBPT187kdI3dMP-ep389egDR2TajYXGG-6xqC2mAlgnBi0vsg=",
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(time.Hour),
		},
	})

	res, err := c.Get(ts.URL + "/public/with-callback")
	require.NoError(t, err)

	assert.EqualValues(t, http.StatusOK, res.StatusCode)
}

func TestIsNotAuthenticated(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	r := x.NewRouterPublic(reg)
	// set this intermediate because kratos needs some valid url for CRUDE operations
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "http://example.com")

	reg.WithCSRFHandler(new(nosurfx.FakeCSRFHandler))
	h, _ := testhelpers.MockSessionCreateHandler(t, reg)
	r.GET("/set", h)
	r.GET("/public/with-callback", reg.SessionHandler().IsNotAuthenticated(send(http.StatusOK), send(http.StatusBadRequest)))
	r.GET("/public/without-callback", reg.SessionHandler().IsNotAuthenticated(send(http.StatusOK), nil))
	ts := httptest.NewServer(r)
	defer ts.Close()

	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, ts.URL)

	sessionClient := testhelpers.NewClientWithCookies(t)
	testhelpers.MockHydrateCookieClient(t, sessionClient, ts.URL+"/set")

	for k, tc := range []struct {
		c    *http.Client
		call string
		code int
	}{
		{
			c:    sessionClient,
			call: "/public/with-callback",
			code: http.StatusBadRequest,
		},
		{
			c:    http.DefaultClient,
			call: "/public/with-callback",
			code: http.StatusOK,
		},

		{
			c:    sessionClient,
			call: "/public/without-callback",
			code: http.StatusForbidden,
		},
		{
			c:    http.DefaultClient,
			call: "/public/without-callback",
			code: http.StatusOK,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			res, err := tc.c.Get(ts.URL + tc.call)
			require.NoError(t, err)

			assert.EqualValues(t, tc.code, res.StatusCode)
		})
	}
}

func TestIsAuthenticated(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	reg.WithCSRFHandler(new(nosurfx.FakeCSRFHandler))
	r := x.NewRouterPublic(reg)

	h, _ := testhelpers.MockSessionCreateHandler(t, reg)
	r.GET("/set", h)
	r.GET("/privileged/with-callback", reg.SessionHandler().IsAuthenticated(send(http.StatusOK), send(http.StatusBadRequest)))
	r.GET("/privileged/without-callback", reg.SessionHandler().IsAuthenticated(send(http.StatusOK), nil))
	ts := httptest.NewServer(r)
	defer ts.Close()
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, ts.URL)

	sessionClient := testhelpers.NewClientWithCookies(t)
	testhelpers.MockHydrateCookieClient(t, sessionClient, ts.URL+"/set")

	for k, tc := range []struct {
		c    *http.Client
		call string
		code int
	}{
		{
			c:    sessionClient,
			call: "/privileged/with-callback",
			code: http.StatusOK,
		},
		{
			c:    http.DefaultClient,
			call: "/privileged/with-callback",
			code: http.StatusBadRequest,
		},

		{
			c:    sessionClient,
			call: "/privileged/without-callback",
			code: http.StatusOK,
		},
		{
			c:    http.DefaultClient,
			call: "/privileged/without-callback",
			code: http.StatusUnauthorized,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			res, err := tc.c.Get(ts.URL + tc.call)
			require.NoError(t, err)

			assert.EqualValues(t, tc.code, res.StatusCode)
		})
	}
}

func TestHandlerAdminSessionManagement(t *testing.T) {
	t.Parallel()

	_, reg := internal.NewFastRegistryWithMocks(t, configx.WithValues(testhelpers.DefaultIdentitySchemaConfig("file://./stub/identity.schema.json")))
	public, ts := testhelpers.NewKratosServer(t, reg)

	t.Run("case=should return 202 after invalidating all sessions", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		var s *Session
		require.NoError(t, faker.FakeData(&s))
		s.Active = true
		s.AMR = AuthenticationMethods{
			{Method: identity.CredentialsTypePassword, CompletedAt: time.Now().UTC().Round(time.Second)},
			{Method: identity.CredentialsTypeOIDC, CompletedAt: time.Now().UTC().Round(time.Second)},
		}
		require.NoError(t, reg.Persister().CreateIdentity(t.Context(), s.Identity))

		var expectedSessionDevice Device
		require.NoError(t, faker.FakeData(&expectedSessionDevice))
		s.Devices = []Device{
			expectedSessionDevice,
		}

		assert.Zero(t, s.ID)
		require.NoError(t, reg.SessionPersister().UpsertSession(t.Context(), s))
		assert.NotZero(t, s.ID)
		assert.NotZero(t, s.Identity.ID)

		t.Run("get session", func(t *testing.T) {
			req, _ := http.NewRequest("GET", ts.URL+"/admin/sessions/"+s.ID.String(), nil)
			res, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			var session Session
			require.NoError(t, json.NewDecoder(res.Body).Decode(&session))
			assert.Equal(t, s.ID, session.ID)
			assert.Nil(t, session.Identity)
			assert.Empty(t, session.Devices)
		})

		t.Run("get session expand", func(t *testing.T) {
			for _, tc := range []struct {
				description        string
				expand             string
				expectedIdentityId string
				expectedDevices    int
			}{
				{
					description:        "expand Identity",
					expand:             "?expand=Identity",
					expectedIdentityId: s.Identity.ID.String(),
					expectedDevices:    0,
				},
				{
					description:        "expand Devices",
					expand:             "/?expand=Devices",
					expectedIdentityId: "",
					expectedDevices:    1,
				},
				{
					description:        "expand Identity and Devices",
					expand:             "/?expand=Identity&expand=Devices",
					expectedIdentityId: s.Identity.ID.String(),
					expectedDevices:    1,
				},
			} {
				t.Run(fmt.Sprintf("description=%s", tc.description), func(t *testing.T) {
					req, _ := http.NewRequest("GET", ts.URL+"/admin/sessions/"+s.ID.String()+tc.expand, nil)
					res, err := client.Do(req)
					require.NoError(t, err)

					body := ioutilx.MustReadAll(res.Body)
					require.Equalf(t, http.StatusOK, res.StatusCode, "%s", body)

					assert.Equal(t, s.ID.String(), gjson.GetBytes(body, "id").String())
					assert.Equal(t, tc.expectedIdentityId, gjson.GetBytes(body, "identity.id").String())
					assert.EqualValuesf(t, tc.expectedDevices, gjson.GetBytes(body, "devices.#").Int(), "%s", gjson.GetBytes(body, "devices").Raw)
				})
			}
		})

		t.Run("get session expand invalid", func(t *testing.T) {
			req, _ := http.NewRequest("GET", ts.URL+"/admin/sessions/"+s.ID.String()+"/?expand=invalid", nil)
			res, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		})

		t.Run("should redirect to public for whoami", func(t *testing.T) {
			client := testhelpers.NewHTTPClientWithSessionToken(t, t.Context(), reg, s)
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			req := testhelpers.NewTestHTTPRequest(t, "GET", ts.URL+"/admin/sessions/whoami", nil)
			res, err := client.Do(req)
			require.NoError(t, err)
			require.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
			require.Equal(t, public.URL+"/sessions/whoami", res.Header.Get("Location"))
		})

		assertPageToken := func(t *testing.T, id, linkHeader string) {
			t.Helper()

			g := link.Parse(linkHeader)
			require.Len(t, g, 1)
			u, err := url.Parse(g["first"].URI)
			require.NoError(t, err)
			pt, err := keysetpagination.NewMapPageToken(u.Query().Get("page_token"))
			require.NoError(t, err)
			mpt := pt.(keysetpagination.MapPageToken)
			assert.Equal(t, id, mpt["id"])
		}

		t.Run("list sessions", func(t *testing.T) {
			req, _ := http.NewRequest("GET", ts.URL+"/admin/sessions/", nil)
			res, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			assertPageToken(t, uuid.Nil.String(), res.Header.Get("Link"))

			var sessions []Session
			require.NoError(t, json.NewDecoder(res.Body).Decode(&sessions))
			require.Len(t, sessions, 1)
			assert.Equal(t, s.ID, sessions[0].ID)
			assert.Empty(t, sessions[0].Identity)
			assert.Empty(t, sessions[0].Devices)
		})

		t.Run("list sessions expand", func(t *testing.T) {
			for _, tc := range []struct {
				description          string
				expand               string
				expectedIdentityId   string
				expectedDevicesCount string
			}{
				{
					description:          "expand nothing",
					expand:               "",
					expectedIdentityId:   "",
					expectedDevicesCount: "",
				},
				{
					description:          "expand Identity",
					expand:               "expand=identity&",
					expectedIdentityId:   s.Identity.ID.String(),
					expectedDevicesCount: "",
				},
				{
					description:          "expand Devices",
					expand:               "expand=devices&",
					expectedIdentityId:   "",
					expectedDevicesCount: "1",
				},
				{
					description:          "expand Identity and Devices",
					expand:               "expand=identity&expand=devices&",
					expectedIdentityId:   s.Identity.ID.String(),
					expectedDevicesCount: "1",
				},
			} {
				t.Run(fmt.Sprintf("description=%s", tc.description), func(t *testing.T) {
					req, _ := http.NewRequest("GET", ts.URL+"/admin/sessions?"+tc.expand, nil)
					res, err := client.Do(req)
					require.NoError(t, err)
					assert.Equal(t, http.StatusOK, res.StatusCode)
					assertPageToken(t, uuid.Nil.String(), res.Header.Get("Link"))

					body := ioutilx.MustReadAll(res.Body)
					assert.Equal(t, s.ID.String(), gjson.GetBytes(body, "0.id").String())
					assert.Equal(t, tc.expectedIdentityId, gjson.GetBytes(body, "0.identity.id").String())
					assert.Equal(t, tc.expectedDevicesCount, gjson.GetBytes(body, "0.devices.#").String())
				})
			}
		})

		t.Run("should list sessions for an identity", func(t *testing.T) {
			req, _ := http.NewRequest("GET", ts.URL+"/admin/identities/"+s.Identity.ID.String()+"/sessions", nil)
			res, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			var sessions []Session
			require.NoError(t, json.NewDecoder(res.Body).Decode(&sessions))
			require.Len(t, sessions, 1)
			assert.Equal(t, s.ID, sessions[0].ID)
		})

		t.Run("should revoke session by id", func(t *testing.T) {
			req, _ := http.NewRequest("GET", ts.URL+"/admin/sessions/"+s.ID.String(), nil)
			res, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			var session Session
			require.NoError(t, json.NewDecoder(res.Body).Decode(&session))
			assert.Equal(t, s.ID, session.ID)
			assert.True(t, session.Active)

			req, _ = http.NewRequest("DELETE", ts.URL+"/admin/sessions/"+s.ID.String(), nil)
			res, err = client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, res.StatusCode)

			req, _ = http.NewRequest("GET", ts.URL+"/admin/sessions/"+s.ID.String(), nil)
			res, err = client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			require.NoError(t, json.NewDecoder(res.Body).Decode(&session))
			assert.Equal(t, s.ID, session.ID)
			assert.False(t, session.Active)
		})

		t.Run("case=session status should be false when session expiry is past", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)

			s.ExpiresAt = time.Now().Add(-time.Hour * 1)
			require.NoError(t, reg.SessionPersister().UpsertSession(t.Context(), s))

			assert.NotEqual(t, uuid.Nil, s.ID)
			assert.NotEqual(t, uuid.Nil, s.Identity.ID)

			req, _ := http.NewRequest("GET", ts.URL+"/admin/sessions/"+s.ID.String(), nil)
			res, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, "false", gjson.GetBytes(body, "active").String(), "%s", body)
		})

		t.Run("case=session status should be false for inactive identity", func(t *testing.T) {
			client := testhelpers.NewClientWithCookies(t)
			var s1 *Session
			require.NoError(t, faker.FakeData(&s1))
			s1.Active = true
			s1.Identity.State = identity.StateInactive
			require.NoError(t, reg.Persister().CreateIdentity(t.Context(), s1.Identity))

			assert.Equal(t, uuid.Nil, s1.ID)
			require.NoError(t, reg.SessionPersister().UpsertSession(t.Context(), s1))
			assert.NotEqual(t, uuid.Nil, s1.ID)
			assert.NotEqual(t, uuid.Nil, s1.Identity.ID)

			req, _ := http.NewRequest("GET", ts.URL+"/admin/sessions/"+s1.ID.String()+"?expand=Identity", nil)
			res, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, "false", gjson.GetBytes(body, "active").String(), "%s", body)
		})

		req, _ := http.NewRequest("DELETE", ts.URL+"/admin/identities/"+s.Identity.ID.String()+"/sessions", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		_, err = reg.SessionPersister().GetSession(t.Context(), s.ID, ExpandNothing)
		require.True(t, errors.Is(err, sqlcon.ErrNoRows))

		t.Run("should not list session", func(t *testing.T) {
			req, _ := http.NewRequest("GET", ts.URL+"/admin/identities/"+s.Identity.ID.String()+"/sessions", nil)
			res, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.JSONEq(t, "[]", string(ioutilx.MustReadAll(res.Body)))
		})
	})

	t.Run("case=should return 400 when bad UUID is sent", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)

		for _, method := range []string{http.MethodGet, http.MethodDelete} {
			t.Run("http method="+method, func(t *testing.T) {
				req, _ := http.NewRequest(method, ts.URL+"/admin/identities/BADUUID/sessions", nil)
				res, err := client.Do(req)
				require.NoError(t, err)
				require.Equal(t, http.StatusBadRequest, res.StatusCode)
			})
		}
	})

	t.Run("case=should return 404 when deleting with unknown UUID", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		someID, _ := uuid.NewV4()
		req, _ := http.NewRequest("DELETE", ts.URL+"/admin/identities/"+someID.String()+"/sessions", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("case=should return pagination headers on list response", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		var i *identity.Identity
		require.NoError(t, faker.FakeData(&i))
		require.NoError(t, reg.Persister().CreateIdentity(t.Context(), i))

		numSessions := 5
		numSessionsActive := 2

		sess := make([]Session, numSessions)
		for j := range sess {
			require.NoError(t, faker.FakeData(&sess[j]))
			sess[j].Identity = i
			if j < numSessionsActive {
				sess[j].Active = true
				sess[j].ExpiresAt = time.Now().UTC().Add(time.Hour)
			} else {
				sess[j].Active = false
				sess[j].ExpiresAt = time.Now().UTC().Add(-time.Hour)
			}
			require.NoError(t, reg.SessionPersister().UpsertSession(t.Context(), &sess[j]))
		}

		for _, tc := range []struct {
			activeOnly         string
			expectedSessionIds []uuid.UUID
		}{
			{
				activeOnly:         "true",
				expectedSessionIds: []uuid.UUID{sess[0].ID, sess[1].ID},
			},
			{
				activeOnly:         "false",
				expectedSessionIds: []uuid.UUID{sess[2].ID, sess[3].ID, sess[4].ID},
			},
			{
				activeOnly:         "",
				expectedSessionIds: []uuid.UUID{sess[0].ID, sess[1].ID, sess[2].ID, sess[3].ID, sess[4].ID},
			},
		} {
			t.Run(fmt.Sprintf("active=%#v", tc.activeOnly), func(t *testing.T) {
				sessions, _, _ := reg.SessionPersister().ListSessionsByIdentity(t.Context(), i.ID, nil, 1, 10, uuid.Nil, ExpandEverything)
				require.Equal(t, 5, len(sessions))
				assert.True(t, sort.IsSorted(sort.Reverse(byCreatedAt(sessions))))

				reqURL := ts.URL + "/admin/identities/" + i.ID.String() + "/sessions"
				if tc.activeOnly != "" {
					reqURL += "?active=" + tc.activeOnly
				}
				req, _ := http.NewRequest("GET", reqURL, nil)
				res, err := client.Do(req)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var actualSessions []Session
				require.NoError(t, json.NewDecoder(res.Body).Decode(&actualSessions))
				actualSessionIds := make([]uuid.UUID, 0)
				for _, s := range actualSessions {
					actualSessionIds = append(actualSessionIds, s.ID)
				}

				assert.NotEqual(t, "", res.Header.Get("Link"))
				assert.ElementsMatch(t, tc.expectedSessionIds, actualSessionIds)
			})
		}
	})
}

func TestHandlerSelfServiceSessionManagement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	ts, _, r, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	// set this intermediate because kratos needs some valid url for CRUDE operations
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "http://example.com")
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, ts.URL)

	var setup func(t *testing.T) (*http.Client, *identity.Identity, *Session)
	{
		// we limit the scope of the channels, so you cannot accidentally mess up a test case
		ident := make(chan *identity.Identity, 1)
		sess := make(chan *Session, 1)
		r.GET("/set", func(w http.ResponseWriter, r *http.Request) {
			h, s := testhelpers.MockSessionCreateHandlerWithIdentity(t, reg, <-ident)
			h(w, r)
			sess <- s
		})

		setup = func(t *testing.T) (*http.Client, *identity.Identity, *Session) {
			client := testhelpers.NewClientWithCookies(t)
			i := identity.NewIdentity("") // the identity is created by the handler

			ident <- i
			testhelpers.MockHydrateCookieClient(t, client, ts.URL+"/set")
			return client, i, <-sess
		}
	}

	t.Run("case=list should return pagination headers", func(t *testing.T) {
		client, i, _ := setup(t)

		numSessions := 5
		numSessionsActive := 2

		sess := make([]Session, numSessions)
		for j := range sess {
			require.NoError(t, faker.FakeData(&sess[j]))
			sess[j].Identity = i
			if j < numSessionsActive {
				sess[j].Active = true
			} else {
				sess[j].Active = false
			}
			require.NoError(t, reg.SessionPersister().UpsertSession(ctx, &sess[j]))
		}

		reqURL := ts.URL + "/sessions"
		req, _ := http.NewRequest("GET", reqURL, nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		require.NotEqual(t, "", res.Header.Get("Link"))
	})

	t.Run("case=should return 200 and number after invalidating all other sessions", func(t *testing.T) {
		client, i, currSess := setup(t)

		otherSess := Session{}
		require.NoError(t, faker.FakeData(&otherSess))
		otherSess.Identity = i
		otherSess.Active = true
		require.NoError(t, reg.SessionPersister().UpsertSession(ctx, &otherSess))

		req, _ := http.NewRequest("DELETE", ts.URL+"/sessions", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)
		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, int64(1), gjson.GetBytes(body, "count").Int(), "%s", body)

		actualOther, err := reg.SessionPersister().GetSession(ctx, otherSess.ID, ExpandNothing)
		require.NoError(t, err)
		assert.False(t, actualOther.Active)

		actualCurr, err := reg.SessionPersister().GetSession(ctx, currSess.ID, ExpandNothing)
		require.NoError(t, err)
		assert.True(t, actualCurr.Active)
	})

	t.Run("case=should revoke specific other session", func(t *testing.T) {
		client, i, _ := setup(t)

		others := make([]Session, 2)
		for j := range others {
			require.NoError(t, faker.FakeData(&others[j]))
			others[j].Identity = i
			others[j].Active = true
			require.NoError(t, reg.SessionPersister().UpsertSession(ctx, &others[j]))
		}

		req, _ := http.NewRequest("DELETE", ts.URL+"/sessions/"+others[0].ID.String(), nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		actualOthers, total, err := reg.SessionPersister().ListSessionsByIdentity(ctx, i.ID, nil, 1, 10, uuid.Nil, ExpandNothing)
		require.NoError(t, err)
		require.Len(t, actualOthers, 3)
		require.Equal(t, int64(3), total)

		for _, s := range actualOthers {
			if s.ID == others[0].ID {
				assert.False(t, s.Active)
			} else {
				assert.True(t, s.Active)
			}
		}
	})

	t.Run("case=should not revoke current session", func(t *testing.T) {
		client, _, currSess := setup(t)

		req, _ := http.NewRequest("DELETE", ts.URL+"/sessions/"+currSess.ID.String(), nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("case=should not error on unknown or revoked session", func(t *testing.T) {
		client, i, _ := setup(t)

		otherSess := Session{}
		require.NoError(t, faker.FakeData(&otherSess))
		otherSess.Identity = i
		otherSess.Active = false
		require.NoError(t, reg.SessionPersister().UpsertSession(ctx, &otherSess))

		for j, id := range []uuid.UUID{otherSess.ID, uuid.Must(uuid.NewV4())} {
			req, _ := http.NewRequest("DELETE", ts.URL+"/sessions/"+id.String(), nil)
			resp, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode, "case=%d", j)
		}
	})

	t.Run("case=whoami should not issue cookie for up to date session", func(t *testing.T) {
		client, _, _ := setup(t)

		req, _ := http.NewRequest("GET", ts.URL+"/sessions/whoami", nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		assert.Empty(t, resp.Cookies())
	})

	t.Run("case=whoami should reissue cookie for outdated session", func(t *testing.T) {
		client, _, session := setup(t)
		oldExpires := session.ExpiresAt

		session.ExpiresAt = time.Now().Add(time.Hour * 24 * 30).UTC().Round(time.Hour)
		err := reg.SessionPersister().UpsertSession(context.Background(), session)
		require.NoError(t, err)

		resp, err := client.Get(ts.URL + "/sessions/whoami")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		require.Len(t, resp.Cookies(), 1)
		for _, c := range resp.Cookies() {
			assert.WithinDuration(t, session.ExpiresAt, c.Expires, 5*time.Second, "Ensure the expiry does not deviate +- 5 seconds from the expiry of the session for cookie: %s", c.Name)
			assert.NotEqual(t, oldExpires, c.Expires, "%s", c.Name)
		}
	})

	t.Run("case=whoami should not issue cookie if request is token based", func(t *testing.T) {
		_, _, session := setup(t)

		session.ExpiresAt = time.Now().Add(time.Hour * 24 * 30).UTC().Round(time.Hour)
		err := reg.SessionPersister().UpsertSession(context.Background(), session)
		require.NoError(t, err)

		req, err := http.NewRequest("GET", ts.URL+"/sessions/whoami", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+session.Token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		require.Len(t, resp.Cookies(), 0)
	})
}

func TestHandlerRefreshSessionBySessionID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	publicServer, adminServer, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	// set this intermediate because kratos needs some valid url for CRUDE operations
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, "http://example.com")
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	conf.MustSet(ctx, config.ViperKeyPublicBaseURL, adminServer.URL)

	i := identity.NewIdentity("")
	require.NoError(t, reg.IdentityManager().Create(context.Background(), i))
	s := &Session{Identity: i, ExpiresAt: time.Now().Add(5 * time.Minute)}
	require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), s))

	t.Run("case=should return 200 after refreshing one session", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)

		req, _ := http.NewRequest("PATCH", adminServer.URL+"/admin/sessions/"+s.ID.String()+"/extend", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		updatedSession, err := reg.SessionPersister().GetSession(context.Background(), s.ID, ExpandNothing)
		require.Nil(t, err)
		require.True(t, s.ExpiresAt.Before(updatedSession.ExpiresAt))
	})

	t.Run("case=should return 400 when bad UUID is sent", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		req, _ := http.NewRequest("PATCH", adminServer.URL+"/admin/sessions/BADUUID/extend", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("case=should return 404 when calling with missing UUID", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		someID, _ := uuid.NewV4()
		req, _ := http.NewRequest("PATCH", adminServer.URL+"/admin/sessions/"+someID.String()+"/extend", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("case=should return 404 when calling puplic server", func(t *testing.T) {
		req := testhelpers.NewTestHTTPRequest(t, "PATCH", publicServer.URL+"/sessions/"+s.ID.String()+"/extend", nil)

		res, err := publicServer.Client().Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
		body := ioutilx.MustReadAll(res.Body)
		assert.NotEqual(t, gjson.GetBytes(body, "error.id").String(), "security_csrf_violation")
	})
}

type byCreatedAt []Session

func (s byCreatedAt) Len() int      { return len(s) }
func (s byCreatedAt) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byCreatedAt) Less(i, j int) bool {
	return s[i].CreatedAt.Before(s[j].CreatedAt)
}
