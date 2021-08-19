package session_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	"github.com/ory/kratos/identity"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/ory/kratos/corpx"
	"github.com/ory/kratos/identity"
	"github.com/ory/x/sqlcon"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	. "github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
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

func assertNoCSRFCookieInResponse(t *testing.T, ts *httptest.Server, c *http.Client, res *http.Response) {
	assert.Len(t, res.Cookies(), 0, res.Cookies())
}

func TestSessionWhoAmI(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	ts, _, r, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	// set this intermediate because kratos needs some valid url for CRUDE operations
	conf.MustSet(config.ViperKeyPublicBaseURL, "http://example.com")
	h, _ := testhelpers.MockSessionCreateHandler(t, reg)
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
			assert.EqualValues(t, ErrAALNotSatisfied.Reason(), gjson.Get(body, "error.reason").String(), body)
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
				assertNoCSRFCookieInResponse(t, ts, client, res) // Test that no CSRF cookie is ever set here.

				assert.EqualValues(t, http.StatusOK, res.StatusCode)
				assert.NotEmpty(t, res.Header.Get("X-Kratos-Authenticated-Identity-Id"))
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
			code: http.StatusForbidden,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			res, err := tc.c.Get(ts.URL + tc.call)
			require.NoError(t, err)

			assert.EqualValues(t, tc.code, res.StatusCode)
		})
	}
}

func TestHandlerDeleteSessionByIdentityID(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	_, ts, _, _ := testhelpers.NewKratosServerWithCSRFAndRouters(t, reg)

	// set this intermediate because kratos needs some valid url for CRUDE operations
	conf.MustSet(config.ViperKeyPublicBaseURL, "http://example.com")
	testhelpers.SetDefaultIdentitySchema(t, conf, "file://./stub/identity.schema.json")
	conf.MustSet(config.ViperKeyPublicBaseURL, ts.URL)

	t.Run("case=should return 202 after invalidating all sessions", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		i := identity.NewIdentity("")
		require.NoError(t, reg.IdentityManager().Create(context.Background(), i))
		s := &Session{Identity: i}
		require.NoError(t, reg.SessionPersister().CreateSession(context.Background(), s))

		req, _ := http.NewRequest("DELETE", ts.URL+"/identities/"+i.ID.String()+"/sessions", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, res.StatusCode)

		_, err = reg.SessionPersister().GetSession(context.Background(), s.ID)
		require.True(t, errors.Is(err, sqlcon.ErrNoRows))
	})

	t.Run("case=should return 400 when bad UUID is sent", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		req, _ := http.NewRequest("DELETE", ts.URL+"/identities/BADUUID/sessions", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("case=should return 404 when calling with missing UUID", func(t *testing.T) {
		client := testhelpers.NewClientWithCookies(t)
		someID, _ := uuid.NewV4()
		req, _ := http.NewRequest("DELETE", ts.URL+"/identities/"+someID.String()+"/sessions", nil)
		res, err := client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}
