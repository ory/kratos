package x

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

// SessionPersistValues adds values to the session store and persists the changes.
func SessionPersistValues(w http.ResponseWriter, r *http.Request, s sessions.StoreExact, id string, values map[string]interface{}) error {
	// The error does not matter because in the worst case we're re-writing the session cookie.
	cookie, _ := s.Get(r, id)
	for k, v := range values {
		cookie.Values[k] = v
	}

	return errors.WithStack(cookie.Save(r, w))
}

// SessionGetString returns a string for the given id and key or an error if the session is invalid,
// the key does not exist, or the key value is not a string.
func SessionGetString(r *http.Request, s sessions.StoreExact, id string, key interface{}) (string, error) {
	check := func(v map[interface{}]interface{}) (string, error) {
		vv, ok := v[key]
		if !ok {
			return "", errors.Errorf("key %s does not exist in cookie: %+v", key, id)
		} else if vvv, ok := vv.(string); !ok {
			return "", errors.Errorf("value of key %s is not of type string in cookie", key)
		} else {
			return vvv, nil
		}
	}

	var exactErr error
	cookie, err := s.GetExact(r, id, func(s *sessions.Session) bool {
		_, exactErr = check(s.Values)
		return exactErr == nil
	})
	if err != nil {
		return "", err
	} else if exactErr != nil {
		return "", exactErr
	}

	return check(cookie.Values)
}

// SessionGetStringOr returns a string for the given id and key or the fallback value if the session is invalid,
// the key does not exist, or the key value is not a string.
func SessionGetStringOr(r *http.Request, s sessions.StoreExact, id, key, fallback string) string {
	v, err := SessionGetString(r, s, id, key)
	if err != nil {
		return fallback
	}
	return v
}

func SessionUnset(w http.ResponseWriter, r *http.Request, s sessions.StoreExact, id string) error {
	cookie, err := s.Get(r, id)
	if err == nil && cookie.IsNew {
		// No cookie was sent in the request. We have nothing to do.
		return nil
	}

	cookie.Options.MaxAge = -1
	cookie.Values = make(map[interface{}]interface{})
	return errors.WithStack(cookie.Save(r, w))
}

func SessionUnsetKey(w http.ResponseWriter, r *http.Request, s sessions.StoreExact, id, key string) error {
	cookie, err := s.Get(r, id)
	if err == nil && cookie.IsNew {
		// No cookie was sent in the request. We have nothing to do.
		return nil
	}

	delete(cookie.Values, key)
	return errors.WithStack(cookie.Save(r, w))
}
