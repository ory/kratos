package x

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

// SessionGetRelayState returns a string of the content of the relaystate for the current session.
func SessionGetStringRelayState(r *http.Request, s sessions.StoreExact, id string, key interface{}) (string, error) {

	cipherRelayState := r.PostForm.Get("RelayState")
	if cipherRelayState == "" {
		return "", errors.New("The RelayState is empty or not exists")
	}

	// Reconstructs the cookie from the ciphered value
	continuityCookie := &http.Cookie{
		Name:   id,
		Value:  cipherRelayState,
		MaxAge: 300,
	}

	r2 := r.Clone(r.Context())
	r2.AddCookie(continuityCookie)

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
	sessionCookie, err := s.GetExact(r2, id, func(s *sessions.Session) bool {
		_, exactErr = check(s.Values)
		return exactErr == nil
	})
	if err != nil {
		return "", err
	} else if exactErr != nil {
		return "", exactErr
	}

	return check(sessionCookie.Values)
}
