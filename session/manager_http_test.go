package session_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ory/nosurf"

	"github.com/ory/kratos/driver"

	"github.com/ory/x/urlx"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/internal/testhelpers"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

var _ nosurf.Handler = new(mockCSRFHandler)

type mockCSRFHandler struct {
	c int
}

func (f *mockCSRFHandler) DisablePath(s string) {
}

func (f *mockCSRFHandler) DisableGlob(s string) {
}

func (f *mockCSRFHandler) DisableGlobs(s ...string) {
}

func (f *mockCSRFHandler) IgnoreGlob(s string) {
}

func (f *mockCSRFHandler) IgnoreGlobs(s ...string) {
}

func (f *mockCSRFHandler) ExemptPath(s string) {}

func (f *mockCSRFHandler) IgnorePath(s string) {}

func (f *mockCSRFHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (f *mockCSRFHandler) RegenerateToken(w http.ResponseWriter, r *http.Request) string {
	f.c++
	return x.FakeCSRFToken
}

func createAAL2Identity(t *testing.T, reg driver.Registry) *identity.Identity {
	idAAL2 := identity.Identity{Traits: []byte("{}"), State: identity.StateActive, Credentials: map[identity.CredentialsType]identity.Credentials{
		identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Config: []byte("{}")},
		identity.CredentialsTypeWebAuthn: {Type: identity.CredentialsTypeWebAuthn, Config: []byte("{}")},
	}}
	return &idAAL2
}

func createAAL1Identity(t *testing.T, reg driver.Registry) *identity.Identity {
	idAAL1 := identity.Identity{Traits: []byte("{}"), State: identity.StateActive, Credentials: map[identity.CredentialsType]identity.Credentials{
		identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Config: []byte("{}")},
	}}
	return &idAAL1
}

func TestManagerHTTP(t *testing.T) {
	t.Run("case=regenerate csrf on principal change", func(t *testing.T) {
		_, reg := internal.NewFastRegistryWithMocks(t)
		mock := new(mockCSRFHandler)
		reg.WithCSRFHandler(mock)

		require.NoError(t, reg.SessionManager().IssueCookie(context.Background(), httptest.NewRecorder(), new(http.Request), new(session.Session)))
		assert.Equal(t, 1, mock.c)
	})

	t.Run("case=cookie settings", func(t *testing.T) {
		ctx := context.Background()
		conf, reg := internal.NewFastRegistryWithMocks(t)
		conf.MustSet("dev", false)
		mock := new(mockCSRFHandler)
		reg.WithCSRFHandler(mock)
		s := &session.Session{Identity: new(identity.Identity)}

		require.NoError(t, conf.Source().Set(config.ViperKeyPublicBaseURL, "https://baseurl.com/base_url"))

		var getCookie = func(t *testing.T, req *http.Request) *http.Cookie {
			rec := httptest.NewRecorder()
			require.NoError(t, reg.SessionManager().IssueCookie(ctx, rec, req, s))
			require.Len(t, rec.Result().Cookies(), 1)
			return rec.Result().Cookies()[0]
		}

		t.Run("case=with default options", func(t *testing.T) {
			actual := getCookie(t, httptest.NewRequest("GET", "https://baseurl.com/bar", nil))
			assert.EqualValues(t, "", actual.Domain, "Domain is empty because unset as a config option")
			assert.EqualValues(t, "/", actual.Path, "Path is the default /")
			assert.EqualValues(t, http.SameSiteLaxMode, actual.SameSite)
			assert.EqualValues(t, true, actual.HttpOnly)
			assert.EqualValues(t, true, actual.Secure)
		})

		t.Run("case=with base cookie customization", func(t *testing.T) {
			conf.MustSet(config.ViperKeyCookiePath, "/cookie")
			conf.MustSet(config.ViperKeyCookieDomain, "cookie.com")
			conf.MustSet(config.ViperKeyCookieSameSite, "Strict")

			actual := getCookie(t, httptest.NewRequest("GET", "https://baseurl.com/bar", nil))
			assert.EqualValues(t, "cookie.com", actual.Domain, "Domain is empty because unset as a config option")
			assert.EqualValues(t, "/cookie", actual.Path, "Path is the default /")
			assert.EqualValues(t, http.SameSiteStrictMode, actual.SameSite)
			assert.EqualValues(t, true, actual.HttpOnly)
			assert.EqualValues(t, true, actual.Secure)
		})

		t.Run("case=with base session customization", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionPath, "/session")
			conf.MustSet(config.ViperKeySessionDomain, "session.com")
			conf.MustSet(config.ViperKeySessionSameSite, "None")

			actual := getCookie(t, httptest.NewRequest("GET", "https://baseurl.com/bar", nil))
			assert.EqualValues(t, "session.com", actual.Domain, "Domain is empty because unset as a config option")
			assert.EqualValues(t, "/session", actual.Path, "Path is the default /")
			assert.EqualValues(t, http.SameSiteNoneMode, actual.SameSite)
			assert.EqualValues(t, true, actual.HttpOnly)
			assert.EqualValues(t, true, actual.Secure)
		})
	})

	t.Run("suite=SessionAddAuthenticationMethod", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/identity.schema.json")

		i := &identity.Identity{Traits: []byte("{}"), State: identity.StateActive}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
		sess := session.NewInactiveSession()
		require.NoError(t, sess.Activate(i, conf, time.Now()))
		require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), sess))
		require.NoError(t, reg.SessionManager().SessionAddAuthenticationMethod(context.Background(), sess.ID, identity.CredentialsTypeOIDC, identity.CredentialsTypeWebAuthn))
		assert.Len(t, sess.AMR, 0)

		actual, err := reg.SessionPersister().GetSession(context.Background(), sess.ID)
		require.NoError(t, err)
		assert.EqualValues(t, identity.AuthenticatorAssuranceLevel2, actual.AuthenticatorAssuranceLevel)
		for _, amr := range actual.AMR {
			assert.True(t, amr.Method == identity.CredentialsTypeWebAuthn || amr.Method == identity.CredentialsTypeOIDC)
		}
		assert.Len(t, actual.AMR, 2)
	})

	t.Run("suite=lifecycle", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		conf.MustSet(config.ViperKeySelfServiceLoginUI, "https://www.ory.sh")
		conf.MustSet(config.ViperKeyDefaultIdentitySchemaURL, "file://./stub/fake-session.schema.json")

		var s *session.Session
		rp := x.NewRouterPublic()
		rp.GET("/session/revoke", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			require.NoError(t, reg.SessionManager().PurgeFromRequest(r.Context(), w, r))
			w.WriteHeader(http.StatusOK)
		})

		rp.GET("/session/set", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			require.NoError(t, reg.SessionManager().UpsertAndIssueCookie(r.Context(), w, r, s))
			w.WriteHeader(http.StatusOK)
		})

		rp.GET("/session/get", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			sess, err := reg.SessionManager().FetchFromRequest(r.Context(), r)
			if err != nil {
				t.Logf("Got error on lookup: %s %T", err, errors.Unwrap(err))
				reg.Writer().WriteError(w, r, err)
				return
			}
			reg.Writer().Write(w, r, sess)
		})

		pts := httptest.NewServer(x.NewTestCSRFHandler(rp, reg))
		t.Cleanup(pts.Close)
		conf.MustSet(config.ViperKeyPublicBaseURL, pts.URL)
		reg.RegisterPublicRoutes(context.Background(), rp)

		t.Run("case=valid", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionLifespan, "1m")

			i := identity.Identity{Traits: []byte("{}")}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, _ = session.NewActiveSession(&i, conf, time.Now(), identity.CredentialsTypePassword)

			c := testhelpers.NewClientWithCookies(t)
			testhelpers.MockHydrateCookieClient(t, c, pts.URL+"/session/set")

			res, err := c.Get(pts.URL + "/session/get")
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=no panic on invalid cookie name", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionLifespan, "1m")
			conf.MustSet(config.ViperKeySessionName, "$%˜\"")
			t.Cleanup(func() {
				conf.MustSet(config.ViperKeySessionName, "")
			})

			rp.GET("/session/set/invalid", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
				require.Error(t, reg.SessionManager().UpsertAndIssueCookie(r.Context(), w, r, s))
				w.WriteHeader(http.StatusInternalServerError)
			})

			i := identity.Identity{Traits: []byte("{}")}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, _ = session.NewActiveSession(&i, conf, time.Now(), identity.CredentialsTypePassword)

			c := testhelpers.NewClientWithCookies(t)
			res, err := c.Get(pts.URL + "/session/set/invalid")
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusInternalServerError, res.StatusCode)
		})

		t.Run("case=valid and uses x-session-cookie", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionLifespan, "1m")

			i := identity.Identity{Traits: []byte("{}")}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, _ = session.NewActiveSession(&i, conf, time.Now(), identity.CredentialsTypePassword)

			c := testhelpers.NewClientWithCookies(t)
			testhelpers.MockHydrateCookieClient(t, c, pts.URL+"/session/set")

			cookies := c.Jar.Cookies(urlx.ParseOrPanic(pts.URL))
			require.Len(t, cookies, 2, "expect two cookies, one csrf, one session")

			var cookie *http.Cookie
			for _, c := range cookies {
				if c.Name == "ory_kratos_session" {
					cookie = c
					break
				}
			}
			require.NotNil(t, cookie, "must find the kratos session cookie")

			assert.Equal(t, "ory_kratos_session", cookie.Name)

			req, err := http.NewRequest("GET", pts.URL+"/session/get", nil)
			require.NoError(t, err)
			req.Header.Set("Cookie", "ory_kratos_session=not-valid")
			req.Header.Set("X-Session-Cookie", cookie.Value)
			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=valid bearer auth as fallback", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionLifespan, "1m")

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

		t.Run("case=valid x-session-token auth even if bearer is set", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionLifespan, "1m")

			i := identity.Identity{Traits: []byte("{}"), State: identity.StateActive}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, err := session.NewActiveSession(&i, conf, time.Now(), identity.CredentialsTypePassword)
			require.NoError(t, err)
			require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), s))

			req, err := http.NewRequest("GET", pts.URL+"/session/get", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer invalid")
			req.Header.Set("X-Session-Token", s.Token)

			c := http.DefaultClient
			res, err := c.Do(req)
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=expired", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionLifespan, "1ns")
			t.Cleanup(func() {
				conf.MustSet(config.ViperKeySessionLifespan, "1m")
			})

			i := identity.Identity{Traits: []byte("{}")}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, _ = session.NewActiveSession(&i, conf, time.Now(), identity.CredentialsTypePassword)

			c := testhelpers.NewClientWithCookies(t)
			testhelpers.MockHydrateCookieClient(t, c, pts.URL+"/session/set")

			time.Sleep(time.Nanosecond * 2)

			res, err := c.Get(pts.URL + "/session/get")
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
		})

		t.Run("case=revoked", func(t *testing.T) {
			i := identity.Identity{Traits: []byte("{}")}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, _ = session.NewActiveSession(&i, conf, time.Now(), identity.CredentialsTypePassword)

			s, _ = session.NewActiveSession(&i, conf, time.Now(), identity.CredentialsTypePassword)

			c := testhelpers.NewClientWithCookies(t)
			testhelpers.MockHydrateCookieClient(t, c, pts.URL+"/session/set")

			res, err := c.Get(pts.URL + "/session/revoke")
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusOK, res.StatusCode)

			res, err = c.Get(pts.URL + "/session/get")
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
		})

		t.Run("case=respects AAL config", func(t *testing.T) {
			conf.MustSet(config.ViperKeySessionLifespan, "1m")

			t.Run("required_aal=aal2", func(t *testing.T) {
				idAAL2 := createAAL2Identity(t, reg)
				idAAL1 := createAAL1Identity(t, reg)
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), idAAL1))
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), idAAL2))

				run := func(t *testing.T, complete []identity.CredentialsType, requested string, i *identity.Identity, expectedError error) {
					s := session.NewInactiveSession()
					for _, m := range complete {
						s.CompletedLoginFor(m)
					}
					require.NoError(t, s.Activate(i, conf, time.Now().UTC()))
					err := reg.SessionManager().DoesSessionSatisfy((&http.Request{}).WithContext(context.Background()), s, requested)
					if expectedError != nil {
						require.ErrorAs(t, err, &expectedError)
					} else {
						require.NoError(t, err)
					}
				}

				t.Run("fulfilled for aal2 if identity has aal2", func(t *testing.T) {
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword, identity.CredentialsTypeWebAuthn}, config.HighestAvailableAAL, idAAL2, nil)
				})

				t.Run("rejected for aal1 if identity has aal2", func(t *testing.T) {
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword}, config.HighestAvailableAAL, idAAL2, session.NewErrAALNotSatisfied(""))
				})

				t.Run("fulfilled for aal1 if identity has aal2 but config is aal1", func(t *testing.T) {
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword}, "aal1", idAAL2, nil)
				})

				t.Run("fulfilled for aal2 if identity has aal1", func(t *testing.T) {
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword}, "aal1", idAAL2, nil)
				})

				t.Run("fulfilled for aal1 if identity has aal1", func(t *testing.T) {
					run(t, []identity.CredentialsType{identity.CredentialsTypePassword}, "aal1", idAAL1, nil)
				})
			})
		})
	})
}
