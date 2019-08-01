package schema

import (
	"strings"

	"github.com/ory/gojsonschema"
)

func ContextSetRoot(c *gojsonschema.JsonContext, head string) *gojsonschema.JsonContext {
	var nc *gojsonschema.JsonContext
	list := append([]string{head}, strings.Split(c.String("."), ".")...)
	for _, item := range list {
		nc = gojsonschema.NewJsonContext(item, nc)
	}
	return nc
}

func ContextRemoveRootStub(c *gojsonschema.JsonContext) *gojsonschema.JsonContext {
	var nc *gojsonschema.JsonContext
	for _, item := range strings.Split(c.String("."), ".") {
		if item != gojsonschema.STRING_CONTEXT_ROOT {
			nc = gojsonschema.NewJsonContext(item, nc)
		}
	}
	if nc == nil {
		return gojsonschema.NewJsonContext("", nil)
	}
	return nc
}
