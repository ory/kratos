package session_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"

	"github.com/tidwall/gjson"

	"github.com/ory/kratos/identity"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/corpx"
	"github.com/ory/x/sqlcon"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	. "github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
	"github.com/ory/x/ioutilx"
	"github.com/ory/x/urlx"
)

func init() {
	corpx.RegisterFakes()
}

func send(code int) httprouter.Handle {
	return func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.WriteHeader(code)
	}
}

func getSessionCookie(t *testing.T, r *http.Response) *http.Cookie {
	var sessionCookie *http.Cookie
	var found bool
	for _, c := range r.Cookies() {
		if c.Name == config.DefaultSessionCookieName {
			found = true
			sessionCookie = c
		}
	}
	require.True(t, found)
	return sessionCookie
}

func assertNoCSRFCookieInResponse(t *testing.T, _ *httptest.Server, _ *http.Client, r *http.Response) {
	found := false
	for _, c := range r.Cookies() {
		if strings.HasPrefix(c.Name, "csrf_token") {
			found = true
		}
	}
	require.False(t, found)
}

func TestSessionWhoAmI(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	ts, _, r, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	// set this intermediate because kratos needs some valid url for CRUDE operations
	conf.MustSet(config.ViperKeyPublicBaseURL, "http://example.com")
	i := &identity.Identity{
		ID:    x.NewUUID(),
		State: identity.StateActive,
		Credentials: map[identity.CredentialsType]identity.Credentials{
			identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword,
				Identifiers: []string{x.NewUUID().String()},
				Config:      []byte(`{"hashed_password":"$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`),
			},
		},
		Traits:         identity.Traits(`{"baz":"bar","foo":true,"bar":2.5}`),
		MetadataAdmin:  []byte(`{"admin":"ma"}`),
		MetadataPublic: []byte(`{"public":"mp"}`),
	}
	h, _ := testhelpers.MockSessionCreateHandlerWithIdentity(t, reg, i)

	r.GET("/set", h)
	conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)

	t.Run("case=aal requirements", func(t *testing.T) {
		h1, _ := testhelpers.MockSessionCreateHandlerWithIdentityAndAMR(t, reg, createAAL2Identity(t, reg), []identity.CredentialsType{identity.CredentialsTypePassword, identity.CredentialsTypeWebAuthn})
		r.GET("/set/aal2-aal2", h1)

		h2, _ := testhelpers.MockSessionCreateHandlerWithIdentityAndAMR(t, reg, createAAL2Identity(t, reg), []identity.CredentialsType{identity.CredentialsTypePassword})
		r.GET("/set/aal2-aal1", h2)

		h3, _ := testhelpers.MockSessionCreateHandlerWithIdentityAndAMR(t, reg, createAAL1Identity(t, reg), []identity.CredentialsType{identity.CredentialsTypePassword})
		r.GET("/set/aal1-aal1", h3)

		run := func(t *testing.T, kind string, code int) string {
			client := testhelpers.NewClientWithCookies(t)
			testhelpers.MockHydrateCookieClient(t, client, ts.URL+"/set/"+kind)

			res, err := client.Get(ts.URL + RouteWhoami)
			require.NoError(t, err)
			body := x.MustReadAll(res.Body)
			assert.EqualValues(t, code, res.StatusCode)
			return string(body)
		}

		t.Run("case=aal2-aal2", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
			run(t, "aal2-aal2", http.StatusOK)
		})

		t.Run("case=aal2-aal2", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionWhoAmIAAL, "aal1")
			run(t, "aal2-aal2", http.StatusOK)
		})

		t.Run("case=aal2-aal1", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
			body := run(t, "aal2-aal1", http.StatusForbidden)
			assert.EqualValues(t, NewErrAALNotSatisfied("").Reason(), gjson.Get(body, "error.reason").String(), body)
		})

		t.Run("case=aal2-aal1", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionWhoAmIAAL, "aal1")
			run(t, "aal2-aal1", http.StatusOK)
		})

		t.Run("case=aal1-aal1", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
			run(t, "aal1-aal1", http.StatusOK)
		})
	})

	t.Run("case=http methods", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)

		// No cookie yet -> 401
		res, err := client.Get(ts.URL + RouteWhoami)
		require.NoError(t, err)
		assertNoCSRFCookieInResponse(t, ts, client, res) // Test that no CSRF cookie is ever set here.

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
				body, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				assertNoCSRFCookieInResponse(t, ts, client, res) // Test that no CSRF cookie is ever set here.

				assert.EqualValues(t, http.StatusOK, res.StatusCode)
				assert.NotEmpty(t, res.Header.Get("X-Kratos-Authenticated-Identity-Id"))

				assert.Empty(t, gjson.GetBytes(body, "identity.credentials"))
				assert.Equal(t, "mp", gjson.GetBytes(body, "identity.metadata_public.public").String(), "%s", body)
				assert.False(t, gjson.GetBytes(body, "identity.metadata_admin").Exists())
			})
		}
	})

	/*
		t.Run("case=respects AAL config", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionLifespan, "1m")

			t.Run("required_aal=aal1", func(t *testing.T) {
				conf.MustSet(config.ViperKeySelfServiceSettingsRequiredAAL, "aal1")

				i := identity.Identity{Traits: []byte("{}"), State: identity.StateActive}
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
				s, err := session.NewActiveSession(&i, conf, time.Now(), identity.CredentialsTypePassword)
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
					conf.MustSet(config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword, identity.CredentialsTypeWebAuthn}, 200, &idAAL2)
				})

				t.Run("rejected for aal1 if identity has aal2", func(t *testing.T) {
					conf.MustSet(config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword}, 403, &idAAL2)
				})

				t.Run("fulfilled for aal1 if identity has aal2 but config is aal1", func(t *testing.T) {
					conf.MustSet(config.ViperKeySessionWhoAmIAAL, "aal1")
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword}, 200, &idAAL2)
				})

				t.Run("fulfilled for aal2 if identity has aal1", func(t *testing.T) {
					conf.MustSet(config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword, identity.CredentialsTypeWebAuthn}, 200, &idAAL1)
				})

				t.Run("fulfilled for aal1 if identity has aal1", func(t *testing.T) {
					conf.MustSet(config.ViperKeySessionWhoAmIAAL, config.HighestAvailableAAL)
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword}, 200, &idAAL1)
				})
			})
		})
	*/
}

func TestIsNotAuthenticatedSecurecookie(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	r := x.NewRouterPublic()
	r.GET("/public/with-callback", reg.SessionHandler().IsNotAuthenticated(send(http.StatusOK), send(http.StatusBadRequest)))

	ts := httptest.NewServer(r)
	defer ts.Close()
	conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)

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
	conf, reg := internal.NewFastRegistryWithMocks(t)
	r := x.NewRouterPublic()
	// set this intermediate because kratos needs some valid url for CRUDE operations
	conf.MustSet(config.ViperKeyPublicBaseURL, "http://example.com")

	reg.WithCSRFHandler(new(x.FakeCSRFHandler))
	h, _ := testhelpers.MockSessionCreateHandler(t, reg)
	r.GET("/set", h)
	r.GET("/public/with-callback", reg.SessionHandler().IsNotAuthenticated(send(http.StatusOK), send(http.StatusBadRequest)))
	r.GET("/public/without-callback", reg.SessionHandler().IsNotAuthenticated(send(http.StatusOK), nil))
	ts := httptest.NewServer(r)
	defer ts.Close()

	conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)

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
	conf, reg := internal.NewFastRegistryWithMocks(t)
	reg.WithCSRFHandler(new(x.FakeCSRFHandler))
	r := x.NewRouterPublic()

	h, _ := testhelpers.MockSessionCreateHandler(t, reg)
	r.GET("/set", h)
	r.GET("/privileged/with-callback", reg.SessionHandler().IsAuthenticated(send(http.StatusOK), send(http.StatusBadRequest)))
	r.GET("/privileged/without-callback", reg.SessionHandler().IsAuthenticated(send(http.StatusOK), nil))
	ts := httptest.NewServer(r)
	defer ts.Close()
	conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)

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
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	_, ts, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	// set this intermediate because kratos needs some valid url for CRUDE operations
	conf.MustSet(config.ViperKeyPublicBaseURL, "http://example.com")
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)

	t.Run("case=should return 202 after invalidating all sessions", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		i := identity.NewIdentity("")
		require.NoError(t, reg.IdentityManager().Create(ctx, i))
		s := &Session{Identity: i}
		require.NoError(t, reg.SessionPersister().UpsertSession(ctx, s))

		t.Run("should list session", func(t *testing.T) {
			req, _ := http.NewRequest("GET", ts.URL+"/admin/identities/"+i.ID.String()+"/sessions", nil)
			res, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			var sessions []Session
			require.NoError(t, json.NewDecoder(res.Body).Decode(&sessions))
			require.Len(t, sessions, 1)
			assert.Equal(t, s.ID, sessions[0].ID)
		})

		req, _ := http.NewRequest("DELETE", ts.URL+"/admin/identities/"+i.ID.String()+"/sessions", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		_, err = reg.SessionPersister().GetSession(ctx, s.ID)
		require.True(t, errors.Is(err, sqlcon.ErrNoRows))

		t.Run("should not list session", func(t *testing.T) {
			req, _ := http.NewRequest("GET", ts.URL+"/admin/identities/"+i.ID.String()+"/sessions", nil)
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

	t.Run("case=should respect active on list", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		i := identity.NewIdentity("")
		require.NoError(t, reg.IdentityManager().Create(ctx, i))

		sess := make([]Session, 2)
		for j := range sess {
			require.NoError(t, faker.FakeData(&sess[j]))
			sess[j].Identity = i
			sess[j].Active = j%2 == 0
			require.NoError(t, reg.SessionPersister().UpsertSession(ctx, &sess[j]))
		}

		for _, tc := range []struct {
			activeOnly  string
			expectedIDs []uuid.UUID
		}{
			{
				activeOnly:  "true",
				expectedIDs: []uuid.UUID{sess[0].ID},
			},
			{
				activeOnly:  "false",
				expectedIDs: []uuid.UUID{sess[1].ID},
			},
			{
				activeOnly:  "",
				expectedIDs: []uuid.UUID{sess[0].ID, sess[1].ID},
			},
		} {
			t.Run(fmt.Sprintf("active=%#v", tc.activeOnly), func(t *testing.T) {
				reqURL := ts.URL + "/admin/identities/" + i.ID.String() + "/sessions"
				if tc.activeOnly != "" {
					reqURL += "?active=" + tc.activeOnly
				}
				req, _ := http.NewRequest("GET", reqURL, nil)
				res, err := client.Do(req)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, res.StatusCode)

				var sessions []Session
				require.NoError(t, json.NewDecoder(res.Body).Decode(&sessions))
				require.Equal(t, len(sessions), len(tc.expectedIDs))

				for _, id := range tc.expectedIDs {
					found := false
					for _, s := range sessions {
						found = found || s.ID == id
					}
					assert.True(t, found)
				}
			})
		}
	})
}

func TestHandlerSelfServiceSessionManagement(t *testing.T) {
	ctx := context.Background()
	conf, reg := internal.NewFastRegistryWithMocks(t)
	ts, _, r, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	// set this intermediate because kratos needs some valid url for CRUDE operations
	conf.MustSet(config.ViperKeyPublicBaseURL, "http://example.com")
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)

	var setup func(t *testing.T) (*http.Client, *identity.Identity, *Session)
	{
		// we limit the scope of the channels, so you cannot accidentally mess up a test case
		ident := make(chan *identity.Identity, 1)
		sess := make(chan *Session, 1)
		r.GET("/set", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			h, s := testhelpers.MockSessionCreateHandlerWithIdentity(t, reg, <-ident)
			h(w, r, ps)
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
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, int64(1), gjson.GetBytes(body, "count").Int(), "%s", body)

		actualOther, err := reg.SessionPersister().GetSession(ctx, otherSess.ID)
		require.NoError(t, err)
		assert.False(t, actualOther.Active)

		actualCurr, err := reg.SessionPersister().GetSession(ctx, currSess.ID)
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

		actualOthers, err := reg.SessionPersister().ListSessionsByIdentity(ctx, i.ID, nil, 1, 10, uuid.Nil)
		require.NoError(t, err)
		require.Len(t, actualOthers, 3)

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

		time.Sleep(500)

		c := reg.Config(context.Background())
		err := reg.SessionPersister().UpsertSession(context.Background(), session.Refresh(c))
		require.NoError(t, err)

		newExpires := session.ExpiresAt
		require.True(t, newExpires.After(oldExpires))

		req, _ := http.NewRequest("GET", ts.URL+"/sessions/whoami", nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		assert.NotEmpty(t, resp.Cookies())
	})
}

func TestHandlerRefreshSessionBySessionID(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	_, ts, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	// set this intermediate because kratos needs some valid url for CRUDE operations
	conf.MustSet(config.ViperKeyPublicBaseURL, "http://example.com")
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")
	conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)

	t.Run("case=should return 200 after refreshing one session", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		i := identity.NewIdentity("")
		require.NoError(t, reg.IdentityManager().Create(context.Background(), i))
		s := &Session{Identity: i, ExpiresAt: time.Now().Add(5 * time.Minute)}
		require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), s))

		req, _ := http.NewRequest("PATCH", ts.URL+"/admin/sessions/"+s.ID.String()+"/extend", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)

		s, err = reg.SessionPersister().GetSession(context.Background(), s.ID)
		require.Nil(t, err)
	})

	t.Run("case=should return 400 when bad UUID is sent", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		req, _ := http.NewRequest("PATCH", ts.URL+"/admin/sessions/BADUUID/extend", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("case=should return 404 when calling with missing UUID", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		someID, _ := uuid.NewV4()
		req, _ := http.NewRequest("PATCH", ts.URL+"/admin/sessions/"+someID.String()+"/extend", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}
