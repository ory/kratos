package password

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/kratos/selfservice"
)

func TestRegistrationRequestMethodConfig(t *testing.T) {
	c := NewRequestMethodConfig()

	assert.Empty(t, c.Fields["foo"].Error)
	c.GetFormFields().SetError("foo", &selfservice.FormError{
		Message: "bar",
	})
	assert.Equal(t, "bar", c.GetFormFields()["foo"].Error.Message)
}
