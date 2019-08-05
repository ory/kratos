package schema

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"

	"github.com/ory/gojsonschema"
)

// NewWindowsCompatibleReferenceLoader returns a JSON reference loader using the given source and the local OS file system.
func NewWindowsCompatibleReferenceLoader(source string) (_ gojsonschema.JSONLoader, err error) {
	if runtime.GOOS == "windows" && strings.HasPrefix(source, "file://") {
		source, err = filepath.Abs(strings.TrimPrefix(source, "file://"))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		source = "file://" + strings.ReplaceAll(source, "\\", "/")
	}
	return gojsonschema.NewReferenceLoader(source), nil
}

func MustNewWindowsCompatibleReferenceLoader(source string) gojsonschema.JSONLoader {
	l, err := NewWindowsCompatibleReferenceLoader(source)
	if err != nil {
		panic(err.Error())
	}
	return l
}
