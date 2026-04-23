package oidc_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type sberMockServer struct {
	server *httptest.Server

	mu          sync.Mutex
	nonceByCode map[string]string
	subByCode   map[string]string
	authDone    int
}

func newSberIFTMockServer(t *testing.T) *sberMockServer {
	t.Helper()

	m := &sberMockServer{
		nonceByCode: make(map[string]string),
		subByCode:   make(map[string]string),
	}

	router := http.NewServeMux()

	router.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		q := r.URL.Query()
		redirectURI := q.Get("redirect_uri")
		state := q.Get("state")
		nonce := q.Get("nonce")
		require.NotEmpty(t, redirectURI)
		require.NotEmpty(t, state)
		require.NotEmpty(t, nonce)

		code := fmt.Sprintf("code-%d", time.Now().UnixNano())
		sub := "sber-ift-user@example.org"

		m.mu.Lock()
		m.nonceByCode[code] = nonce
		m.subByCode[code] = sub
		m.mu.Unlock()

		target, err := url.Parse(redirectURI)
		require.NoError(t, err)
		query := target.Query()
		query.Set("code", code)
		query.Set("state", state)
		target.RawQuery = query.Encode()

		http.Redirect(w, r, target.String(), http.StatusSeeOther)
	})

	router.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, r.ParseForm())

		code := r.Form.Get("code")
		clientID := r.Form.Get("client_id")
		require.NotEmpty(t, code)
		require.NotEmpty(t, clientID)

		m.mu.Lock()
		nonce := m.nonceByCode[code]
		sub := m.subByCode[code]
		m.mu.Unlock()
		require.NotEmpty(t, nonce)
		require.NotEmpty(t, sub)

		now := time.Now().Unix()
		idToken := makeUnsignedJWT(map[string]any{
			"aud":   clientID,
			"sub":   sub,
			"nonce": nonce,
			"iat":   now,
			"exp":   now + 300,
		})

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access-" + code,
			"token_type":    "bearer",
			"expires_in":    300,
			"id_token":      idToken,
			"refresh_token": "refresh-" + code,
		}))
	})

	router.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "Bearer access-"))
		require.NotEmpty(t, r.Header.Get("x-introspect-rquid"))

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"sub":          "sber-ift-user@example.org",
			"email":        "sber-ift-user@example.org",
			"phone_number": "+7 (999) 123-45-67",
			"given_name":   "ИВАН",
			"family_name":  "ИВАНОВ",
			"middle_name":  "ИВАНОВИЧ",
			"birthdate":    "1990-01-02",
			"gender":       2,
		}))
	})

	router.HandleFunc("/auth/completed", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.True(t, strings.HasPrefix(r.Header.Get("authorization"), "Bearer access-"))
		require.NotEmpty(t, r.Header.Get("rquid"))
		m.mu.Lock()
		m.authDone++
		m.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	m.server = httptest.NewServer(router)
	t.Cleanup(m.server.Close)

	return m
}

func (m *sberMockServer) URL() string {
	return m.server.URL
}

func (m *sberMockServer) AuthCompletedCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.authDone
}

func makeUnsignedJWT(claims map[string]any) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payloadRaw, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(payloadRaw)
	return header + "." + payload + ".sig"
}
