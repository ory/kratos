package schema

import (
	"strings"

	"github.com/ory/jsonschema/v3"
)

func ContextSetRoot(err *jsonschema.ValidationError, head string) *jsonschema.ValidationError {
	err.InstancePtr = prefixPointer(err.InstancePtr, head)

	if err.Context != nil {
		switch ctx := err.Context.(type) {
		case *jsonschema.ValidationErrorContextRequired:
			for k, pointer := range ctx.Missing {
				ctx.Missing[k] = prefixPointer(pointer, head)
			}
		}
	}

	for k := range err.Causes {
		err.Causes[k] = ContextSetRoot(err.Causes[k], head)
	}

	return err
}

func prefixPointer(pointer, head string) string {
	if pointer == "#" || pointer == "#/" {
		pointer = "#/" + head
	} else {
		pointer = "#/" + head + "/" + strings.TrimPrefix(pointer, "#/")
	}
	return pointer
}
