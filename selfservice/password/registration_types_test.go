package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistrationRequestMethodConfig(t *testing.T) {
	c := NewRegistrationRequestMethodConfig()

	assert.Empty(t, c.Fields["foo"].Error)
	c.GetFormFields().SetError("foo", "bar")
	assert.Equal(t, "bar", c.GetFormFields()["foo"].Error)
}
