package selfservice

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormFields(t *testing.T) {
	t.Run("method=SetValue", func(t *testing.T) {
		ff := FormFields{
			"1": {Name: "1", Value: "foo"},
			"2": {Name: "2", Value: ""},
		}
		assert.Len(t, ff, 2)

		ff.SetValue("1", "baz1")
		ff.SetValue("2", "baz2")
		ff.SetValue("3", "baz3")

		assert.Len(t, ff, 3)
		for _, k := range []string{"1", "2", "3"} {
			assert.EqualValues(t, fmt.Sprintf("baz%s", k), ff[k].Value, "%+v", ff)
		}
	})

	t.Run("method=AddError", func(t *testing.T) {
		ff := FormFields{
			"1": {Name: "1", Value: "foo", Error: &FormError{Message: "foo"}},
			"2": {Name: "2", Value: "", Error: &FormError{Message: ""}},
		}
		assert.Len(t, ff, 2)

		ff.SetError("1", &FormError{Message: "baz1"})
		ff.SetError("2", &FormError{Message: "baz2"})
		ff.SetError("3", &FormError{Message: "baz3"})

		assert.Len(t, ff, 3)
		for _, k := range []string{"1", "2", "3"} {
			assert.EqualValues(t, fmt.Sprintf("baz%s", k), ff[k].Error.Message, "%+v", ff)
		}
	})

	t.Run("method=Reset", func(t *testing.T) {
		ff := FormFields{
			"1": {Name: "1", Value: "foo", Error: &FormError{Message: "foo"}},
			"2": {Name: "2", Value: "bar", Error: &FormError{Message: "bar"}},
		}

		ff.Reset()

		assert.Empty(t, ff["1"].Error)
		assert.Empty(t, ff["1"].Value)

		assert.Empty(t, ff["2"].Error)
		assert.Empty(t, ff["2"].Value)
	})
}
