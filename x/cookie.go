package x

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

// SessionPersistValues adds values to the session store and persists the changes.
func SessionPersistValues(w http.ResponseWriter, r *http.Request, s sessions.Store, id string, values map[string]interface{}) error {
	// The error does not matter because in the worst case we're re-writing the session cookie.
	cookie, err := s.Get(r, id)
	if err != nil {
		cookie = sessions.NewSession(s, id)
	}

	for k, v := range values {
		cookie.Values[k] = v
	}

	return errors.WithStack(cookie.Save(r, w))
}

// SessionGetString returns a string for the given id and key or an error if the session is invalid,
// the key does not exist, or the key value is not a string.
func SessionGetString(r *http.Request, s sessions.Store, id string, key interface{}) (string, error) {
	cookie, err := s.Get(r, id)
	if err != nil {
		return "", errors.WithStack(err)
	}

	if v, ok := cookie.Values[key]; !ok {
		return "", errors.Errorf("key %s does not exist in cookie", key)
	} else if vv, ok := v.(string); !ok {
		return "", errors.Errorf("value of key %s is not of type string in cookie", key)
	} else {
		return vv, nil
	}
}

// SessionGetStringOr returns a string for the given id and key or the fallback value if the session is invalid,
// the key does not exist, or the key value is not a string.
func SessionGetStringOr(r *http.Request, s sessions.Store, id, key, fallback string) string {
	v, err := SessionGetString(r, s, id, key)
	if err != nil {
		return fallback
	}
	return v
}
