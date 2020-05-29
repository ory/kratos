package questions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("method=recovery", func(t *testing.T) {
		t.SkipNow()
		assert.EqualValues(t, []RecoverySecurityQuestion{{ID: "foo", Label: "bar"}},
			nil)
		// p.SelfServiceRecoverySecurityQuestions())
	})
}
