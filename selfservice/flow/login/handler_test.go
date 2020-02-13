package login_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/gofrs/uuid"
	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/kratos/selfservice/form"
	"github.com/ory/kratos/selfservice/strategy/password"

	"github.com/ory/viper"

	"github.com/ory/kratos/driver/configuration"
	"github.com/ory/kratos/internal"
	"github.com/ory/kratos/selfservice/errorx"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/selfservice/strategy/oidc"
	"github.com/ory/kratos/session"
	"github.com/ory/kratos/x"
)

type (
	withTokenGenerator interface {
		WithTokenGenerator(g form.CSRFGenerator)
	}
	testHTTPWriter struct {
		*testing.T
		r *http.Request
	}
	testHTTPHeader struct {
		http.Header
		r *http.Request
		t *testing.T
	}
)

func (h testHTTPHeader) Add(key, value string) {
	if h.r != nil && key == "Set-Cookie" {
		header := http.Header{}
		header.Add(key, value)
		req := http.Request{Header: header}
		for _, c := range req.Cookies() {
			h.r.AddCookie(c)
		}
	} else {
		h.t.Logf("Add() request: %+v", h.r)
	}

	h.t.Logf("called Add(%s, %s)\n", key, value)
}

func (t testHTTPWriter) Header() http.Header {
	t.Log("called Header()\n")
	return http.Header(testHTTPHeader{
		Header: http.Header{},
		r:      t.r,
		t:      t.T,
	})
}

func (t testHTTPWriter) Write(b []byte) (int, error) {
	t.Logf("called Write(\"%s\")\n", b)
	return len(b), nil
}

func (t testHTTPWriter) WriteHeader(statusCode int) {
	t.Logf("called WriteHeader(%d)\n", statusCode)
}

func init() {
	internal.RegisterFakes()
}

func TestSessionReauthFunctionality(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)
	for _, strategy := range reg.LoginStrategies() {
		// We need to know the csrf token
		strategy.(withTokenGenerator).WithTokenGenerator(x.FakeCSRFTokenGenerator)
	}

	router := x.NewRouterPublic()
	admin := x.NewRouterAdmin()
	reg.LoginHandler().RegisterPublicRoutes(router)
	reg.LoginHandler().RegisterAdminRoutes(admin)
	reg.LoginHandler().WithTokenGenerator(x.FakeCSRFTokenGenerator)
	reg.SelfServiceErrorManager().WithTokenGenerator(x.FakeCSRFTokenGenerator)
	reg.LoginStrategies().RegisterPublicRoutes(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("already authenticated"))
	}))
	defer redirTS.Close()

	loginTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := r.URL.Query().Get("request")

		if request != "" {
			_, err := w.Write([]byte(request))
			require.NoError(t, err)
			return
		}
		_, err := w.Write([]byte("no request id"))
		require.NoError(t, err)
	}))

	errorTS := errorx.NewErrorTestServer(t, reg)

	viper.Set(configuration.ViperKeyURLsDefaultReturnTo, redirTS.URL)
	viper.Set(configuration.ViperKeyURLsSelfPublic, ts.URL)
	viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)
	viper.Set(configuration.ViperKeyURLsError, errorTS.URL)
	viper.Set(configuration.ViperKeyDefaultIdentityTraitsSchemaURL, "file://./stub/login.schema.json")

	pathWithRequestQuery := func(path, otherQuery string) func(id uuid.UUID) (string, string) {
		return func(id uuid.UUID) (string, string) {
			qs := fmt.Sprintf("request=%s", id)
			if otherQuery != "" {
				qs += "&" + otherQuery
			}
			return path, qs
		}
	}
	oidcAuthPath := func(id uuid.UUID) (string, string) {
		return strings.ReplaceAll(oidc.AuthPath, ":request", id.String()), ""
	}
	for k, tc := range []struct {
		method        string
		pathAndQuery  func(id uuid.UUID) (string, string)
		formContent   url.Values
		modifyRequest func(t *testing.T, r *http.Request, rid string)
	}{
		{
			method:       "GET",
			pathAndQuery: pathWithRequestQuery(login.BrowserLoginPath, ""),
		},
		{
			method:       "POST",
			pathAndQuery: pathWithRequestQuery(password.LoginPath, ""),
			formContent: url.Values{
				"identifier": {"foo"},
				"password":   {"zq32klsjd12ed"},
			},
		},
		{
			method:       "POST",
			pathAndQuery: oidcAuthPath,
		},
		{
			method:       "GET",
			pathAndQuery: oidcAuthPath,
		},
		{
			method:       "GET",
			pathAndQuery: pathWithRequestQuery(oidc.CallbackPath, "state=state"),
			modifyRequest: func(t *testing.T, r *http.Request, rid string) {
				require.NoError(t, x.SessionPersistValues(testHTTPWriter{t, r}, r, reg.CookieManager(), "oicd_session", map[string]interface{}{
					"state":      "state",
					"request_id": rid,
					"form":       r.PostForm.Encode(),
				}))
				r.AddCookie(sessions.NewCookie("state", "state", &sessions.Options{
					Path:     "/",
					HttpOnly: true,
				}))
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d/method=%s", k, tc.method), func(t *testing.T) {
			rid := x.NewUUID()
			path, qs := tc.pathAndQuery(rid)

			req := x.NewTestHTTPRequest(t, tc.method, ts.URL+path, bytes.NewBuffer(([]byte)(tc.formContent.Encode())))
			loginReq := login.NewLoginRequest(time.Minute, x.FakeCSRFToken, req)
			loginReq.ID = rid
			for _, s := range reg.LoginStrategies() {
				require.NoError(t, s.PopulateLoginMethod(req, loginReq))
			}
			require.NoError(t, reg.LoginRequestPersister().CreateLoginRequest(context.TODO(), loginReq), loginReq)

			req.URL.RawQuery = qs
			if tc.modifyRequest != nil {
				tc.modifyRequest(t, req, rid.String())
			}

			body, resp := session.MockMakeAuthenticatedRequest(t, reg, router.Router, req)
			assert.EqualValues(t, resp.Request.URL.Query().Get("request"), string(body), resp)
			assert.True(t, strings.HasPrefix(resp.Request.URL.String(), loginTS.URL+"?request="))
		})
	}
}

func TestLoginHandler(t *testing.T) {
	_, reg := internal.NewRegistryDefault(t)

	public, admin := func() (*httptest.Server, *httptest.Server) {
		public := x.NewRouterPublic()
		admin := x.NewRouterAdmin()
		reg.LoginHandler().RegisterPublicRoutes(public)
		reg.LoginHandler().RegisterAdminRoutes(admin)
		reg.LoginStrategies().RegisterPublicRoutes(public)
		return httptest.NewServer(nosurf.New(public)), httptest.NewServer(admin)
	}()
	defer public.Close()

	redirTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer redirTS.Close()

	easyGet := func(t *testing.T, c *http.Client, url string) []byte {
		res, err := c.Get(url)
		require.NoError(t, err)
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		return body
	}

	newLoginTS := func(t *testing.T, upstream string, c *http.Client) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c == nil {
				c = http.DefaultClient
			}
			_, err := w.Write(easyGet(t, c, upstream+login.BrowserLoginRequestsPath+"?request="+r.URL.Query().Get("request")))
			require.NoError(t, err)
		}))
	}

	assertRequestPayload := func(t *testing.T, body []byte) {
		assert.Equal(t, "password", gjson.GetBytes(body, "methods.password.method").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "methods.password.config.fields.#(name==csrf_token).value").String(), "%s", body)
		assert.NotEmpty(t, gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Empty(t, gjson.GetBytes(body, "headers").Value(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.password.config.action").String(), gjson.GetBytes(body, "id").String(), "%s", body)
		assert.Contains(t, gjson.GetBytes(body, "methods.password.config.action").String(), public.URL, "%s", body)
	}

	errTS := errorx.NewErrorTestServer(t, reg)
	defer errTS.Close()

	viper.Set(configuration.ViperKeyURLsSelfPublic, public.URL)
	viper.Set(configuration.ViperKeyURLsError, errTS.URL)

	t.Run("daemon=admin", func(t *testing.T) {
		loginTS := newLoginTS(t, admin.URL, nil)
		defer loginTS.Close()

		viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)
		assertRequestPayload(t, easyGet(t, public.Client(), public.URL+login.BrowserLoginPath))
	})

	t.Run("daemon=public", func(t *testing.T) {
		j, err := cookiejar.New(nil)
		require.NoError(t, err)
		hc := &http.Client{Jar: j}

		t.Run("case=with_csrf", func(t *testing.T) {
			loginTS := newLoginTS(t, public.URL, hc)
			defer loginTS.Close()
			viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)

			assertRequestPayload(t, easyGet(t, hc, public.URL+login.BrowserLoginPath))
		})

		t.Run("case=without_csrf", func(t *testing.T) {
			loginTS := newLoginTS(t, public.URL,
				// using a different client because it doesn't have access to the cookie jar
				new(http.Client))
			defer loginTS.Close()
			viper.Set(configuration.ViperKeyURLsLogin, loginTS.URL)

			body := easyGet(t, hc, public.URL+login.BrowserLoginPath)
			assert.Contains(t, gjson.GetBytes(body, "error").String(), "csrf_token", "%s", body)
		})
	})
}
