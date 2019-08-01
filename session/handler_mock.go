package session

import (
	"net/http"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"

	"github.com/ory/hive/identity"
)

func MockSetSession(t *testing.T, reg Registry) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, err := reg.SessionManager().CreateToRequest(&identity.Identity{}, w, r)
		require.NoError(t, err)
		w.WriteHeader(http.StatusNoContent)
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
