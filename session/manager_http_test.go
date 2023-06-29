// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package session_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ory/nosurf"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver"

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
		identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Config: []byte(`{"hashed_password": "$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`), Identifiers: []string{testhelpers.RandomEmail()}},
		identity.CredentialsTypeWebAuthn: {Type: identity.CredentialsTypeWebAuthn, Config: []byte(`{"credentials":[{"is_passwordless":false}]}`), Identifiers: []string{testhelpers.RandomEmail()}},
	}}
	return &idAAL2
}

func createAAL1Identity(t *testing.T, reg driver.Registry) *identity.Identity {
	idAAL1 := identity.Identity{Traits: []byte("{}"), State: identity.StateActive, Credentials: map[identity.CredentialsType]identity.Credentials{
		identity.CredentialsTypePassword: {Type: identity.CredentialsTypePassword, Config: []byte(`{"hashed_password": "$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`), Identifiers: []string{testhelpers.RandomEmail()}},
	}}
	return &idAAL1
}

func TestManagerHTTP(t *testing.T) {
	ctx := context.Background()

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
		conf.MustSet(ctx, "dev", false)
		mock := new(mockCSRFHandler)
		reg.WithCSRFHandler(mock)
		s := &session.Session{Identity: new(identity.Identity)}

		require.NoError(t, conf.GetProvider(ctx).Set(config.ViperKeyPublicBaseURL, "https://baseurl.com/base_url"))

		getCookie := func(t *testing.T, req *http.Request) *http.Cookie {
			rec := httptest.NewRecorder()
			require.NoError(t, reg.SessionManager().IssueCookie(ctx, rec, req, s))
			require.Len(t, rec.Result().Cookies(), 1)
			return rec.Result().Cookies()[0]
		}

		t.Run("case=immutability", func(t *testing.T) {
			cookie1 := getCookie(t, x.NewTestHTTPRequest(t, "GET", "https://baseurl.com/bar", nil))
			cookie2 := getCookie(t, x.NewTestHTTPRequest(t, "GET", "https://baseurl.com/bar", nil))

			assert.NotEqual(t, cookie1.Value, cookie2.Value)
		})

		t.Run("case=with default options", func(t *testing.T) {
			actual := getCookie(t, httptest.NewRequest("GET", "https://baseurl.com/bar", nil))
			assert.EqualValues(t, "", actual.Domain, "Domain is empty because unset as a config option")
			assert.EqualValues(t, "/", actual.Path, "Path is the default /")
			assert.EqualValues(t, http.SameSiteLaxMode, actual.SameSite)
			assert.EqualValues(t, true, actual.HttpOnly)
			assert.EqualValues(t, true, actual.Secure)
		})

		t.Run("case=with base cookie customization", func(t *testing.T) {
			conf.MustSet(ctx, config.ViperKeyCookiePath, "/cookie")
			conf.MustSet(ctx, config.ViperKeyCookieDomain, "cookie.com")
			conf.MustSet(ctx, config.ViperKeyCookieSameSite, "Strict")

			actual := getCookie(t, httptest.NewRequest("GET", "https://baseurl.com/bar", nil))
			assert.EqualValues(t, "cookie.com", actual.Domain, "Domain is empty because unset as a config option")
			assert.EqualValues(t, "/cookie", actual.Path, "Path is the default /")
			assert.EqualValues(t, http.SameSiteStrictMode, actual.SameSite)
			assert.EqualValues(t, true, actual.HttpOnly)
			assert.EqualValues(t, true, actual.Secure)
		})

		t.Run("case=with base session customization", func(t *testing.T) {
			conf.MustSet(ctx, config.ViperKeySessionPath, "/session")
			conf.MustSet(ctx, config.ViperKeySessionDomain, "session.com")
			conf.MustSet(ctx, config.ViperKeySessionSameSite, "None")

			actual := getCookie(t, httptest.NewRequest("GET", "https://baseurl.com/bar", nil))
			assert.EqualValues(t, "session.com", actual.Domain, "Domain is empty because unset as a config option")
			assert.EqualValues(t, "/session", actual.Path, "Path is the default /")
			assert.EqualValues(t, http.SameSiteNoneMode, actual.SameSite)
			assert.EqualValues(t, true, actual.HttpOnly)
			assert.EqualValues(t, true, actual.Secure)
		})
	})

	t.Run("suite=SessionAddAuthenticationMethod", func(t *testing.T) {
		req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)

		conf, reg := internal.NewFastRegistryWithMocks(t)
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")

		i := &identity.Identity{Traits: []byte("{}"), State: identity.StateActive}
		require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), i))
		sess := session.NewInactiveSession()
		require.NoError(t, sess.Activate(req, i, conf, time.Now()))
		require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), sess))
		require.NoError(t, reg.SessionManager().SessionAddAuthenticationMethods(context.Background(), sess.ID,
			session.AuthenticationMethod{
				Method: identity.CredentialsTypeOIDC,
				AAL:    identity.AuthenticatorAssuranceLevel1,
			},
			session.AuthenticationMethod{
				Method: identity.CredentialsTypeWebAuthn,
				AAL:    identity.AuthenticatorAssuranceLevel2,
			}))
		assert.Len(t, sess.AMR, 0)

		actual, err := reg.SessionPersister().GetSession(context.Background(), sess.ID, session.ExpandNothing)
		require.NoError(t, err)
		assert.EqualValues(t, identity.AuthenticatorAssuranceLevel2, actual.AuthenticatorAssuranceLevel)
		for _, amr := range actual.AMR {
			assert.True(t, amr.Method == identity.CredentialsTypeWebAuthn || amr.Method == identity.CredentialsTypeOIDC)
		}
		assert.Len(t, actual.AMR, 2)
	})

	t.Run("suite=lifecycle", func(t *testing.T) {
		conf, reg := internal.NewFastRegistryWithMocks(t)
		conf.MustSet(ctx, config.ViperKeySelfServiceLoginUI, "https://www.ory.sh")
		testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/fake-session.schema.json")

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
		conf.MustSet(ctx, config.ViperKeyPublicBaseURL, pts.URL)
		reg.RegisterPublicRoutes(context.Background(), rp)

		t.Run("case=valid", func(t *testing.T) {
			req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
			conf.MustSet(req.Context(), config.ViperKeySessionLifespan, "1m")

			i := identity.Identity{Traits: []byte("{}")}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, _ = session.NewActiveSession(req, &i, conf, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)

			c := testhelpers.NewClientWithCookies(t)
			testhelpers.MockHydrateCookieClient(t, c, pts.URL+"/session/set")

			res, err := c.Get(pts.URL + "/session/get")
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=key rotation", func(t *testing.T) {
			req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
			original := conf.GetProvider(ctx).Strings(config.ViperKeySecretsCookie)
			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeySecretsCookie, original)
			})
			conf.MustSet(ctx, config.ViperKeySessionLifespan, "1m")
			conf.MustSet(ctx, config.ViperKeySecretsCookie, []string{"foo"})

			i := identity.Identity{Traits: []byte("{}")}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, _ = session.NewActiveSession(req, &i, conf, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)

			c := testhelpers.NewClientWithCookies(t)
			testhelpers.MockHydrateCookieClient(t, c, pts.URL+"/session/set")

			res, err := c.Get(pts.URL + "/session/get")
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusOK, res.StatusCode)

			conf.MustSet(ctx, config.ViperKeySecretsCookie, []string{"bar", "foo"})
			res, err = c.Get(pts.URL + "/session/get")
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=no panic on invalid cookie name", func(t *testing.T) {
			req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
			conf.MustSet(ctx, config.ViperKeySessionLifespan, "1m")
			conf.MustSet(ctx, config.ViperKeySessionName, "$%˜\"")
			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeySessionName, "")
			})

			rp.GET("/session/set/invalid", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
				require.Error(t, reg.SessionManager().UpsertAndIssueCookie(r.Context(), w, r, s))
				w.WriteHeader(http.StatusInternalServerError)
			})

			i := identity.Identity{Traits: []byte("{}")}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, _ = session.NewActiveSession(req, &i, conf, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)

			c := testhelpers.NewClientWithCookies(t)
			res, err := c.Get(pts.URL + "/session/set/invalid")
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusInternalServerError, res.StatusCode)
		})

		t.Run("case=valid bearer auth as fallback", func(t *testing.T) {
			req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
			conf.MustSet(ctx, config.ViperKeySessionLifespan, "1m")

			i := identity.Identity{Traits: []byte("{}"), State: identity.StateActive}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, err := session.NewActiveSession(req, &i, conf, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
			require.NoError(t, err)
			require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), s))
			require.NotEmpty(t, s.Token)

			req, err = http.NewRequest("GET", pts.URL+"/session/get", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+s.Token)

			c := http.DefaultClient
			res, err := c.Do(req)
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=valid x-session-token auth even if bearer is set", func(t *testing.T) {
			req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
			conf.MustSet(ctx, config.ViperKeySessionLifespan, "1m")

			i := identity.Identity{Traits: []byte("{}"), State: identity.StateActive}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, err := session.NewActiveSession(req, &i, conf, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)
			require.NoError(t, err)
			require.NoError(t, reg.SessionPersister().UpsertSession(context.Background(), s))

			req, err = http.NewRequest("GET", pts.URL+"/session/get", nil)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer invalid")
			req.Header.Set("X-Session-Token", s.Token)

			c := http.DefaultClient
			res, err := c.Do(req)
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusOK, res.StatusCode)
		})

		t.Run("case=expired", func(t *testing.T) {
			req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
			conf.MustSet(ctx, config.ViperKeySessionLifespan, "1ns")
			t.Cleanup(func() {
				conf.MustSet(ctx, config.ViperKeySessionLifespan, "1m")
			})

			i := identity.Identity{Traits: []byte("{}")}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, _ = session.NewActiveSession(req, &i, conf, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)

			c := testhelpers.NewClientWithCookies(t)
			testhelpers.MockHydrateCookieClient(t, c, pts.URL+"/session/set")

			time.Sleep(time.Nanosecond * 2)

			res, err := c.Get(pts.URL + "/session/get")
			require.NoError(t, err)
			assert.EqualValues(t, http.StatusUnauthorized, res.StatusCode)
		})

		t.Run("case=revoked", func(t *testing.T) {
			req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
			i := identity.Identity{Traits: []byte("{}")}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), &i))
			s, _ = session.NewActiveSession(req, &i, conf, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)

			s, _ = session.NewActiveSession(req, &i, conf, time.Now(), identity.CredentialsTypePassword, identity.AuthenticatorAssuranceLevel1)

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
			conf.MustSet(ctx, config.ViperKeySessionLifespan, "1m")

			t.Run("required_aal=aal2", func(t *testing.T) {
				req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
				idAAL2 := createAAL2Identity(t, reg)
				idAAL1 := createAAL1Identity(t, reg)
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), idAAL1))
				require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), idAAL2))

				run := func(t *testing.T, complete []identity.CredentialsType, requested string, i *identity.Identity, expectedError error) {
					s := session.NewInactiveSession()
					for _, m := range complete {
						s.CompletedLoginFor(m, "")
					}
					require.NoError(t, s.Activate(req, i, conf, time.Now().UTC()))
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

func TestDoesSessionSatisfy(t *testing.T) {
	conf, reg := internal.NewFastRegistryWithMocks(t)
	testhelpers.SetDefaultIdentitySchema(conf, "file://./stub/identity.schema.json")

	password := identity.Credentials{
		Type:        identity.CredentialsTypePassword,
		Identifiers: []string{testhelpers.RandomEmail()},
		Config:      []byte(`{"hashed_password": "$argon2id$v=19$m=32,t=2,p=4$cm94YnRVOW5jZzFzcVE4bQ$MNzk5BtR2vUhrp6qQEjRNw"}`),
	}
	oidc := identity.Credentials{
		Type:        identity.CredentialsTypeOIDC,
		Config:      []byte(`{"providers":[{"subject":"0.fywegkf7hd@ory.sh","provider":"hydra","initial_id_token":"65794a68624763694f694a53557a49314e694973496d74705a434936496e4231596d7870597a706f6557527959533576634756756157517561575174644739725a5734694c434a30655841694f694a4b5631516966512e65794a686446396f59584e6f496a6f6956484650616b6f324e6c397a613046436555643662315679576b466655534973496d46315a43493657794a72636d463062334d74593278705a573530496c3073496d46316447686664476c745a5349364d5459304e6a55314e6a59784e4377695a586877496a6f784e6a51324e5459774d6a45314c434a70595851694f6a45324e4459314e5459324d545573496d6c7a63794936496d6830644841364c79397362324e6862476876633351364e4451304e4338694c434a7164476b694f694a6a596a4d784d6a51794e6930314e7a4d774c5451314d546374596a51335a53316b4d446379596a51334d6a6b344d4759694c434a79595851694f6a45324e4459314e5459324d544d73496e4e705a434936496a677a4e5755344e47526a4c5463344d544d744e4749324f4330354d544a6d4c5446684d7a646d4e444d354d4463304e534973496e4e3159694936496a41755a6e6c335a5764725a6a646f5a454276636e6b75633267694c434a335a574a7a6158526c496a6f696148523063484d364c7939336433637562334a354c6e4e6f4c794a392e506850623770456358544c3456647730427959686f30794a7232714b794b4f7373646c4b6c74716b4953693762414e58776a7635686538506e6d7a586e713538556f5739657754584a485a33425651614d4e79612d755f5933584a4a61665673543347476c52776f376f5261707a6a564836502d72447657385649524d5361356f783242397164416d796659505734376e56782d4e68787247564c56464b526b5866324e4448534e6d435968524963455539724331366235385331344c314367776972624d507662797870644c63764f4a4546554238324c794574525a786f644748354c69394d6b5f4d6137363969583254776758434179306734475a625957337137317466574c37736d5342394669785076434b6a3738433753546b762d764f737a4e6533523864676133775471466e6253797a6a614f4b47626e424a4a77423869306e416c48496d425337587146645f666d556d4e62377a372d63716e593374395069306248466b46596e6746545279664d4c6f466f576956784842704b4d6c6b304d4e7a5155414e5368546e346769544d5547454a4f6372346f6f445f6770344768734c44542d54465f6f73486c304832544237777a6d546d735f3150506547424e716a316b61576a467038567247726e4a6b354f594c643152473152464c794535544c4d47315f62744762447137334450784c334b3657387348507242504b654133344377373371584e5247724e73574e69496e775f4e596a65554d484b6351436c4e51445a49725339794962456a485a78476a34546e4367664f5974694e76527a4c6c36616a73614265464b7a45592d6348416e6e42694c75744439373168697241684f5463544a42783672716f67717764755356726551456f565a5735616e4a7a7575775234685453354d44314d64457045437471526d416c71555459644e5a365778514d","initial_access_token":"52344752743736552d634a2d4a2d424372447159634967464652446c6455455a6a526e534d62336e3242732e47324f444d64303544774b4e67395649476e306e496b3877324e72444f48384a78635042635a4a58336d63","initial_refresh_token":"327872337a4d382d654273674b6d61644a624e5a497572473374545154615070313264514a314476544d632e77326d34747a6e7950584c38324b794563716468685068635156314f77386a535a345355496f3544744a51"}]}`),
		Identifiers: []string{"hydra:0.fywegkf7hd@ory.sh"},
	}
	mfaWebAuth := identity.Credentials{
		Type:        identity.CredentialsTypeWebAuthn,
		Config:      []byte(`{"credentials":[{"is_passwordless":false}]}`),
		Identifiers: []string{testhelpers.RandomEmail()},
	}
	passwordlessWebAuth := identity.Credentials{
		Type:        identity.CredentialsTypeWebAuthn,
		Config:      []byte(`{"credentials":[{"is_passwordless":true}]}`),
		Identifiers: []string{testhelpers.RandomEmail()},
	}
	webAuthEmpty := identity.Credentials{Type: identity.CredentialsTypeWebAuthn, Config: []byte(`{}`), Identifiers: []string{testhelpers.RandomEmail()}}
	passwordEmpty := identity.Credentials{Type: identity.CredentialsTypePassword, Config: []byte(`{}`), Identifiers: []string{testhelpers.RandomEmail()}}

	amrPassword := session.AuthenticationMethod{Method: identity.CredentialsTypePassword, AAL: identity.AuthenticatorAssuranceLevel1}

	for k, tc := range []struct {
		d                     string
		err                   error
		requested             identity.AuthenticatorAssuranceLevel
		creds                 []identity.Credentials
		amr                   session.AuthenticationMethods
		sessionManagerOptions []session.ManagerOptions
		expectedFunc          func(t *testing.T, err error, tcError error)
	}{
		{
			d:         "has=aal1, requested=highest, available=aal1, credential=password",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password},
			amr:       session.AuthenticationMethods{amrPassword},
		},
		{
			d:         "has=aal1, requested=highest, available=aal1, credential=password, legacy=true",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password},
			amr:       session.AuthenticationMethods{{Method: identity.CredentialsTypePassword}},
		},
		{
			d:         "has=aal1, requested=highest, available=aal1, credential=password+webauth_empty",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password, webAuthEmpty},
			amr:       session.AuthenticationMethods{amrPassword},
		},
		{
			d:         "has=aal1, requested=highest, available=aal1, credential=password+webauth_empty, legacy=true",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password, webAuthEmpty},
			amr:       session.AuthenticationMethods{{Method: identity.CredentialsTypePassword}},
		},
		{
			d:         "has=aal1, requested=highest, available=aal1, credential=password+webauth_passwordless",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password, passwordlessWebAuth},
			amr:       session.AuthenticationMethods{amrPassword},
		},
		{
			d:         "has=aal1, requested=highest, available=aal1, credential=password+webauth_passwordless, legacy=true",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password, passwordlessWebAuth},
			amr:       session.AuthenticationMethods{{Method: identity.CredentialsTypePassword}},
		},
		{
			d:         "has=aal1, requested=highest, available=aal2, credential=password+webauth_mfa",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password, mfaWebAuth},
			amr:       session.AuthenticationMethods{amrPassword},
			err:       new(session.ErrAALNotSatisfied),
		},
		{
			d:         "has=aal1, requested=highest, available=aal2, credential=password+webauth_mfa, legacy=true",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password, mfaWebAuth},
			amr:       session.AuthenticationMethods{{Method: identity.CredentialsTypePassword}},
			err:       new(session.ErrAALNotSatisfied),
		},
		{
			d:         "has=aal1, requested=highest, available=aal2, credential=password+webauth_mfa",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password, mfaWebAuth},
			amr:       session.AuthenticationMethods{amrPassword, {Method: identity.CredentialsTypeWebAuthn, AAL: identity.AuthenticatorAssuranceLevel1}},
			err:       new(session.ErrAALNotSatisfied),
		},
		{
			d:         "has=aal1, requested=highest, available=aal2, credential=password+webauth_passwordless",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password, passwordlessWebAuth},
			amr:       session.AuthenticationMethods{amrPassword, {Method: identity.CredentialsTypeWebAuthn, AAL: identity.AuthenticatorAssuranceLevel1}},
		},
		{
			d:         "has=aal2, requested=highest, available=aal2, credential=password+webauth_mfa",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password, mfaWebAuth},
			amr:       session.AuthenticationMethods{amrPassword, {Method: identity.CredentialsTypeWebAuthn, AAL: identity.AuthenticatorAssuranceLevel2}},
		},
		{
			d:         "has=aal2, requested=highest, available=aal2, credential=password+webauth_mfa, legacy=true",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password, mfaWebAuth},
			amr:       session.AuthenticationMethods{amrPassword, {Method: identity.CredentialsTypeWebAuthn}},
		},
		{
			d:         "has=aal1, requested=highest, available=aal1, credential=oidc_and_empties",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{oidc, webAuthEmpty, passwordEmpty},
			amr:       session.AuthenticationMethods{{Method: identity.CredentialsTypeOIDC, AAL: identity.AuthenticatorAssuranceLevel1}},
		},
		{
			d:                     "has=aal1, requested=highest, available=aal1, credentials=password+webauthn_mfa, recovery with session manager options",
			requested:             config.HighestAvailableAAL,
			creds:                 []identity.Credentials{password, mfaWebAuth},
			amr:                   session.AuthenticationMethods{{Method: identity.CredentialsTypeRecoveryCode}},
			err:                   session.NewErrAALNotSatisfied(urlx.CopyWithQuery(urlx.AppendPaths(conf.SelfPublicURL(context.Background()), "/self-service/login/browser"), url.Values{"aal": {"aal2"}, "return_to": {"https://myapp.com/settings?id=123"}}).String()),
			sessionManagerOptions: []session.ManagerOptions{session.WithRequestURL("https://myapp.com/settings?id=123")},
			expectedFunc: func(t *testing.T, err error, tcError error) {
				require.Contains(t, err.(*session.ErrAALNotSatisfied).RedirectTo, "myapp.com")
				require.Equal(t, tcError.(*session.ErrAALNotSatisfied).RedirectTo, err.(*session.ErrAALNotSatisfied).RedirectTo)
			},
		},
		{
			d:         "has=aal1, requested=highest, available=aal1, credentials=password+webauthn_mfa, recovery without session manager options",
			requested: config.HighestAvailableAAL,
			creds:     []identity.Credentials{password, mfaWebAuth},
			amr:       session.AuthenticationMethods{{Method: identity.CredentialsTypeRecoveryCode}},
			err:       session.NewErrAALNotSatisfied(urlx.CopyWithQuery(urlx.AppendPaths(conf.SelfPublicURL(context.Background()), "/self-service/login/browser"), url.Values{"aal": {"aal2"}}).String()),
			expectedFunc: func(t *testing.T, err error, tcError error) {
				require.Equal(t, tcError.(*session.ErrAALNotSatisfied).RedirectTo, err.(*session.ErrAALNotSatisfied).RedirectTo)
			},
		},
	} {
		t.Run(fmt.Sprintf("run=%d/desc=%s", k, tc.d), func(t *testing.T) {
			id := identity.NewIdentity("")
			for _, c := range tc.creds {
				id.SetCredentials(c.Type, c)
			}
			require.NoError(t, reg.PrivilegedIdentityPool().CreateIdentity(context.Background(), id))
			t.Cleanup(func() {
				require.NoError(t, reg.PrivilegedIdentityPool().DeleteIdentity(context.Background(), id.ID))
			})

			req := x.NewTestHTTPRequest(t, "GET", "/sessions/whoami", nil)
			s := session.NewInactiveSession()
			for _, m := range tc.amr {
				s.CompletedLoginFor(m.Method, m.AAL)
			}
			require.NoError(t, s.Activate(req, id, conf, time.Now().UTC()))

			err := reg.SessionManager().DoesSessionSatisfy((&http.Request{}).WithContext(context.Background()), s, string(tc.requested), tc.sessionManagerOptions...)
			if tc.err != nil {
				if tc.expectedFunc != nil {
					tc.expectedFunc(t, err, tc.err)
				}

				require.ErrorAs(t, err, &tc.err)

			} else {
				require.NoError(t, err)
			}

			// This should still work even if the session does not have identity data attached yet...
			s.Identity = nil
			err = reg.SessionManager().DoesSessionSatisfy((&http.Request{}).WithContext(context.Background()), s, string(tc.requested), tc.sessionManagerOptions...)
			if tc.err != nil {
				if tc.expectedFunc != nil {
					tc.expectedFunc(t, err, tc.err)
				}
				require.ErrorAs(t, err, &tc.err)
			} else {
				require.NoError(t, err)
			}

			// ..or no credentials attached.
			s.Identity = id
			s.Identity.Credentials = nil
			err = reg.SessionManager().DoesSessionSatisfy((&http.Request{}).WithContext(context.Background()), s, string(tc.requested), tc.sessionManagerOptions...)
			if tc.err != nil {
				if tc.expectedFunc != nil {
					tc.expectedFunc(t, err, tc.err)
				}
				require.ErrorAs(t, err, &tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
