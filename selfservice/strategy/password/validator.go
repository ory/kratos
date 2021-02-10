package password

import (
	"bufio"
	"context"

	"github.com/ory/kratos/driver/config"

	/* #nosec G505 sha1 is used for k-anonymity */
	"crypto/sha1"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/arbovm/levenshtein"

	"github.com/ory/x/httpx"

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
	Validate(ctx context.Context, identifier, password string) error
}

type ValidationProvider interface {
	PasswordValidator() Validator
}

var _ Validator = new(DefaultPasswordValidator)
var ErrNetworkFailure = errors.New("unable to check if password has been leaked because an unexpected network error occurred")
var ErrUnexpectedStatusCode = errors.New("unexpected status code")

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
	reg    validatorDependencies
	Client *http.Client
	hashes map[string]int64

	minIdentifierPasswordDist            int
	maxIdentifierPasswordSubstrThreshold float32
}

type validatorDependencies interface {
	config.Provider
}

func NewDefaultPasswordValidatorStrategy(reg validatorDependencies) *DefaultPasswordValidator {
	return &DefaultPasswordValidator{
		Client:                    httpx.NewResilientClientLatencyToleranceMedium(nil),
		reg:                       reg,
		hashes:                    map[string]int64{},
		minIdentifierPasswordDist: 5, maxIdentifierPasswordSubstrThreshold: 0.5}
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

func (s *DefaultPasswordValidator) fetch(hpw []byte, apiDNSName string) error {
	prefix := fmt.Sprintf("%X", hpw)[0:5]
	loc := fmt.Sprintf("https://%s/range/%s", apiDNSName, prefix)
	res, err := s.Client.Get(loc)
	if err != nil {
		return errors.Wrapf(ErrNetworkFailure, "%s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Wrapf(ErrUnexpectedStatusCode, "%d", res.StatusCode)
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

func (s *DefaultPasswordValidator) Validate(ctx context.Context, identifier, password string) error {
	if len(password) < 6 {
		return errors.Errorf("password length must be at least 6 characters but only got %d", len(password))
	}

	compIdentifier, compPassword := strings.ToLower(identifier), strings.ToLower(password)
	dist := levenshtein.Distance(compIdentifier, compPassword)
	lcs := float32(lcsLength(compIdentifier, compPassword)) / float32(len(compPassword))
	if dist < s.minIdentifierPasswordDist || lcs > s.maxIdentifierPasswordSubstrThreshold {
		return errors.Errorf("the password is too similar to the user identifier")
	}

	/* #nosec G401 sha1 is used for k-anonymity */
	h := sha1.New()
	if _, err := h.Write([]byte(password)); err != nil {
		return err
	}
	hpw := h.Sum(nil)

	s.RLock()
	c, ok := s.hashes[b20(hpw)]
	s.RUnlock()

	passwordPolicyConfig := s.reg.Config(ctx).PasswordPolicyConfig()

	if !ok {
		err := s.fetch(hpw, passwordPolicyConfig.HaveIBeenPwnedHost)
		if (errors.Is(err, ErrNetworkFailure) || errors.Is(err, ErrUnexpectedStatusCode)) && passwordPolicyConfig.IgnoreNetworkErrors {
			return nil
		} else if err != nil {
			return err
		}

		return s.Validate(ctx, identifier, password)
	}

	if c > int64(s.reg.Config(ctx).PasswordPolicyConfig().MaxBreaches) {
		return errors.Errorf("the password has been found in at least %d data breaches and must no longer be used.", c)
	}

	return nil
}
