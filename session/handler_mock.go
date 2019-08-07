package session

import (
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/bxcodec/faker"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/hive/identity"
)

func MockSetSession(t *testing.T, reg Registry) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, err := reg.SessionManager().CreateToRequest(&identity.Identity{}, w, r)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}
}

func MockGetSession(t *testing.T, reg Registry) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, err := reg.SessionManager().FetchFromRequest(r)
		if r.URL.Query().Get("has") == "yes" {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func MockMakeAuthenticatedRequest(t *testing.T, reg Registry, router *httprouter.Router, req *http.Request) ([]byte, *http.Response) {
	set := "/" + uuid.New().String() + "/set"
	router.GET(set, MockSetSession(t, reg))

	client := MockCookieClient(t)
	MockHydrateCookieClient(t, client, "http://"+req.URL.Host+set)

	res, err := client.Do(req)
	require.NoError(t, err)

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	require.NoError(t, res.Body.Close())

	return body, res
}

func MockCookieClient(t *testing.T) *http.Client {
	cj, err := cookiejar.New(&cookiejar.Options{})
	require.NoError(t, err)
	return &http.Client{Jar: cj}
}

func MockHydrateCookieClient(t *testing.T, c *http.Client, u string) {
	res, err := c.Get(u)
	require.NoError(t, err)
	assert.EqualValues(t, http.StatusOK, res.StatusCode)

	var found bool
	for _, c := range res.Cookies() {
		if c.Name == DefaultSessionCookieName {
			found = true
		}
	}
	require.True(t, found)
}

func MockSessionCreateHandler(t *testing.T, reg Registry) httprouter.Handle {
	var sess Session
	require.NoError(t, faker.FakeData(&sess))
	require.NoError(t, reg.SessionManager().Create(&sess))

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		require.NoError(t, reg.SessionManager().Create(&sess))
		require.NoError(t, reg.SessionManager().SaveToRequest(&sess, w, r))
	}
}
