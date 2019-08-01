package x

import (
	errors2 "errors"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestFormError(t *testing.T) {
	t.Run("stackTrace", func(t *testing.T) {
		err1 := errors.New("asdf")
		err2 := (&FormError{}).WithError(err1)
		assert.Equal(t, err1.(stackTracer).StackTrace(), err2.StackTrace())
		assert.Nil(t, (&FormError{}).WithError(errors2.New("asdf")).StackTrace())
		assert.Nil(t, (&FormError{}).StackTrace())
	})

	t.Run("fields", func(t *testing.T) {
		err := &FormError{}
		assert.Empty(t, err.MessageField)

		err2 := err.WithMessage("message")
		assert.Equal(t, "message", err2.MessageField)
		assert.NotEqual(t, err.MessageField, err2.MessageField)

		err2 = err.WithMessagef("message %s", "message")
		assert.Equal(t, "message message", err2.MessageField)
		assert.NotEqual(t, err.MessageField, err2.MessageField)

		err2 = err.WithField("field")
		assert.Equal(t, "field", err2.FieldField)
		assert.NotEqual(t, err.FieldField, err2.FieldField)

		err2 = err.WithError(errors.New("err"))
		assert.Equal(t, "err", err2.Error())
		assert.NotEqual(t, err.Error(), err2.Error())
	})
}
