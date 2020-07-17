package password

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLCSLength(t *testing.T) {
	for k, tc := range []struct {
		a string
		b string
		l int
	}{
		{a: "foo", b: "foo", l: 3},
		{a: "fo", b: "foo", l: 2},
		{a: "bar", b: "foo", l: 0},
		{a: "foobar", b: "foo", l: 3},
		{a: "foobar", b: "oo", l: 2},
		{a: "foobar", b: "a", l: 1},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			require.Equal(t, tc.l, lcsLength(tc.a, tc.b))
			require.Equal(t, tc.l, lcsLength(tc.b, tc.a))
		})
	}
}

func TestDefaultPasswordValidationStrategy(t *testing.T) {
	// Tests are based on:
	// - https://www.troyhunt.com/passwords-evolved-authentication-guidance-for-the-modern-era/
	// - https://www.microsoft.com/en-us/research/wp-content/uploads/2016/06/Microsoft_Password_Guidance-1.pdf

	s := NewDefaultPasswordValidatorStrategy()
	for k, tc := range []struct {
		id   string
		pw   string
		pass bool
	}{
		{pw: "", pass: false},
		{pw: "12", pass: false},
		{pw: "1234", pass: false},
		{pw: "123456", pass: false},
		{pw: "12345678", pass: false},
		{pw: "password", pass: false},
		{pw: "1234567890", pass: false},
		{pw: "qwertyui", pass: false},
		{pw: "l3f9toh1uaf81n21", pass: true},
		{pw: "l3f9toh1uaf81n21", id: "l3f9toh1uaf81n21", pass: false},
		{pw: "l3f9toh1", pass: true},
		{pw: "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", pass: true},
		// simple permutation tests
		{id: "hello@example.com", pw: "hello@example.com12345", pass: false},
		{id: "hello@example.com", pw: "123hello@example.com123", pass: false},
		{id: "hello@example.com", pw: "hello@exam", pass: false},
		{id: "hello@example.com", pw: "HELLO@EXAMPLE.COM", pass: false},
		{id: "hello@example.com", pw: "HELLO@example.com", pass: false},
		{pw: "hello@example.com", id: "hello@exam", pass: false},
		{id: "hello@example.com", pw: "h3ll0@example", pass: false},
		{pw: "hello@example.com", id: "hello@exam", pass: false},
		{id: "abcd", pw: "9d3c8a1b", pass: true},
		{id: "a", pw: "kjOkla", pass: true},
		{id: "ab", pw: "0000ab0000", pass: true},
		// longest common substring with long password
		{id: "d4f6090b-5a84", pw: "d4f6090b-5a84-2184-4404-8d1b-8da3eb00ebbe", pass: true},
		{id: "asdflasdflasdf", pw: "asdflasdflpiuhefnciluaksdzuföfhg", pass: true},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			err := s.Validate(tc.id, tc.pw)
			if tc.pass {
				require.NoError(t, err, "err: %+v, id: %s, pw: %s", err, tc.id, tc.pw)
			} else {
				require.Error(t, err, "id: %s, pw: %s", tc.id, tc.pw)
			}
		})
	}
}
