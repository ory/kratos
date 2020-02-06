package password

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/arbovm/levenshtein"

	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/stringsx"
)

// Validator implements a validation strategy for passwords. One example is that the password
// has to have at least 6 characters and at least one lower and one uppercase password.
type Validator interface {
	// Validate returns nil if the password is passing the validation strategy and an error otherwise. If a validation error
	// occurs, a regular error will be returned. If some other type of error occurs (e.g. HTTP request failed), an error
	// of type *herodot.DefaultError will be returned.
	Validate(identifier, password string) error
}

type ValidationProvider interface {
	PasswordValidator() Validator
}

var _ Validator = new(DefaultPasswordValidator)

// DefaultPasswordValidator implements Validator. It is based on best
// practices as defined in the following blog posts:
//
// - https://www.troyhunt.com/passwords-evolved-authentication-guidance-for-the-modern-era/
// - https://www.microsoft.com/en-us/research/wp-content/uploads/2016/06/Microsoft_Password_Guidance-1.pdf
//
// Additionally passwords are being checked against Troy Hunt's
// [haveibeenpwnd](https://haveibeenpwned.com/API/v2#SearchingPwnedPasswordsByRange) service to check if the
// password has been breached in a previous data leak using k-anonymity.
type DefaultPasswordValidator struct {
	sync.RWMutex
	c      *http.Client
	hashes map[string]int64

	maxBreachesThreshold int64
	ignoreNetworkErrors  bool

	minIdentifierPasswordDist   int
	maxIdentifierPasswordSubstr int
}

func NewDefaultPasswordValidatorStrategy() *DefaultPasswordValidator {
	return &DefaultPasswordValidator{
		c:                           http.DefaultClient,
		maxBreachesThreshold:        0,
		hashes:                      map[string]int64{},
		ignoreNetworkErrors:         true,
		minIdentifierPasswordDist:   5,
		maxIdentifierPasswordSubstr: 3,
	}
}

func NewDefaultPasswordValidatorStrategyStrict() *DefaultPasswordValidator {
	v := NewDefaultPasswordValidatorStrategy()
	v.ignoreNetworkErrors = false
	return v
}

func b20(src []byte) string {
	return fmt.Sprintf("%X", src)
}

// code inspired by https://rosettacode.org/wiki/Longest_Common_Substring#Go
func lcsLength(a, b string) int {
	lengths := make([]int, len(a)*len(b))
	greatestLength := 0
	for i, x := range a {
		for j, y := range b {
			if x == y {
				curr := 1
				if i != 0 && j != 0 {
					curr = lengths[(i-1)*len(b)+j-1] + 1
				}

				if curr > greatestLength {
					greatestLength = curr
				}
				lengths[i*len(b)+j] = curr
			}
		}
	}
	return greatestLength
}

func (s *DefaultPasswordValidator) fetch(hpw []byte) error {
	prefix := fmt.Sprintf("%X", hpw)[0:5]
	loc := fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix)
	res, err := s.c.Get(loc)
	if err != nil {
		if s.ignoreNetworkErrors {
			return nil
		}
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to check if password has been breached before: %s", err))
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if s.ignoreNetworkErrors {
			return nil
		}
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to check if password has been breached before, expected status code 200 but got %d", res.StatusCode))
	}

	s.Lock()
	s.hashes[b20(hpw)] = 0
	s.Unlock()

	sc := bufio.NewScanner(res.Body)
	for sc.Scan() {
		row := sc.Text()
		result := stringsx.Splitx(strings.TrimSpace(row), ":")

		if len(result) != 2 {
			return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected password hash from remote to contain two parts separated by a double dot but got: %v (%s)", result, row))
		}

		count, err := strconv.ParseInt(result[1], 10, 64)
		if err != nil {
			return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected password hash to contain a count formatted as int but got: %s", result[1]))
		}

		s.Lock()
		s.hashes[(prefix + result[0])] = count
		s.Unlock()
	}

	if err := sc.Err(); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to initialize string scanner: %s", err))
	}

	return nil
}

func (s *DefaultPasswordValidator) Validate(identifier, password string) error {
	if len(password) < 6 {
		return errors.Errorf("password length must be at least 6 characters but only got %d", len(password))
	}

	if levenshtein.Distance(identifier, password) < s.minIdentifierPasswordDist || lcsLength(identifier, password) > s.maxIdentifierPasswordSubstr {
		return errors.Errorf("the password is to similar to the user identifier")
	}

	h := sha1.New()
	if _, err := h.Write([]byte(password)); err != nil {
		return err
	}
	hpw := h.Sum(nil)

	s.RLock()
	c, ok := s.hashes[b20(hpw)]
	s.RUnlock()

	if !ok {
		if err := s.fetch(hpw); err != nil {
			return err
		}

		return s.Validate(identifier, password)
	}

	if c > s.maxBreachesThreshold {
		return errors.Errorf("the password has been found in at least %d data breaches and must no longer be used.", c)
	}

	return nil
}
