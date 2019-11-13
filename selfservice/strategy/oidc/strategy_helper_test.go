package oidc_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/resilience"
	"github.com/ory/x/urlx"

	"github.com/ory/kratos/driver"
	"github.com/ory/kratos/x"
)

func newErrTs(t *testing.T, reg driver.Registry) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e, err := reg.ErrorManager().Read(r.Context(), r.URL.Query().Get("error"))
		require.NoError(t, err)
		reg.Writer().Write(w, r, e)
	}))
}

func createClient(t *testing.T, remote string, redir string) {
	require.NoError(t, resilience.Retry(logrus.New(), time.Second*10, time.Minute*2, func() error {
		if req, err := http.NewRequest("DELETE", remote+"/clients/client", nil); err != nil {
			return err
		} else if _, err := http.DefaultClient.Do(req); err != nil {
			return err
		}

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(&struct {
			ClientID      string   `json:"client_id"`
			ClientSecret  string   `json:"client_secret"`
			Scope         string   `json:"scope"`
			GrantTypes    []string `json:"grant_types"`
			ResponseTypes []string `json:"response_types"`
			RedirectURIs  []string `json:"redirect_uris"`
		}{
			ClientID:      "client",
			ClientSecret:  "secret",
			GrantTypes:    []string{"authorization_code", "refresh_token"},
			ResponseTypes: []string{"code"},
			Scope:         "offline offline_access openid",
			RedirectURIs:  []string{redir},
		}))

		res, err := http.Post(remote+"/clients", "application/json", &b)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if http.StatusCreated != res.StatusCode {
			return errors.Errorf("got status code: %d", http.StatusCreated)
		}
		return nil
	}))
}

func newHydraIntegration(t *testing.T, remote *string, subject *string, scope *[]string, addr string) (*http.Server, string) {
	router := httprouter.New()

	type p struct {
		Subject    string   `json:"subject,omitempty"`
		GrantScope []string `json:"grant_scope,omitempty"`
	}

	var do = func(w http.ResponseWriter, r *http.Request, href string, payload io.Reader) {
		req, err := http.NewRequest("PUT", href, payload)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		body := x.MustReadAll(res.Body)
		require.Equal(t, http.StatusOK, res.StatusCode, "%s", body)

		var response struct {
			RedirectTo string `json:"redirect_to"`
		}
		require.NoError(t, json.NewDecoder(bytes.NewBuffer(body)).Decode(&response))
		require.NotNil(t, response.RedirectTo, "%s", body)

		http.Redirect(w, r, response.RedirectTo, http.StatusFound)
	}

	router.GET("/login", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NotEmpty(t, *remote)
		require.NotEmpty(t, *subject)

		challenge := r.URL.Query().Get("login_challenge")
		require.NotEmpty(t, challenge)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(&p{Subject: *subject}))
		href := urlx.MustJoin(*remote, "/oauth2/auth/requests/login/accept") + "?login_challenge=" + challenge
		do(w, r, href, &b)
	})

	router.GET("/consent", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NotEmpty(t, *remote)
		require.NotNil(t, *scope)

		challenge := r.URL.Query().Get("consent_challenge")
		require.NotEmpty(t, challenge)

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(&p{GrantScope: *scope}))
		href := urlx.MustJoin(*remote, "/oauth2/auth/requests/consent/accept") + "?consent_challenge=" + challenge
		do(w, r, href, &b)
	})

	if addr == "" {
		server := httptest.NewServer(router)
		return server.Config, server.URL
	}
	server := &http.Server{Addr: addr, Handler: router}
	go func() {
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
		} else if err != nil {
			panic(err)
		}
	}()
	return server, fmt.Sprintf("http://%s", addr)
}

func newReturnTs(t *testing.T, reg driver.Registry) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := reg.SessionManager().FetchFromRequest(r.Context(), w, r)
		require.NoError(t, err)
		require.Empty(t, sess.Identity.Credentials)
		reg.Writer().Write(w, r, sess)
	}))
}
