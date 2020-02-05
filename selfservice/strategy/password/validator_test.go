package password

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

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
		{pw: "l3f9toh1uaf81n21", id: "l3f9txh1uaf81n21", pass: false},
		{pw: "l3f9toh1uaf81n21", id: "l3f9txh1uafa1n21", pass: false},
		{pw: "l3f9toh1uaf81n21", id: "l3f9txh1uafa1n11", pass: false},
		{pw: "l3f9toh1uaf81n21", id: "q3f9txh1uafa1n11", pass: false},
		{pw: "l3f9toh1uaf81n21", id: "q3x9txh1uafa1n11", pass: true},
		{pw: "l3f9toh1", pass: true},
		{pw: "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", pass: true},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			err := s.Validate(tc.id, tc.pw)
			if tc.pass {
				require.NoError(t, err, "%+v", err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
