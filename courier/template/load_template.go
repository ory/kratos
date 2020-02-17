package template

import (
	"bytes"
	"io"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/gobuffalo/packr/v2"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

var box = packr.New("templates", "templates")
var cache, _ = lru.New(16)

func loadTextTemplate(path string, model interface{}) (string, error) {
	var b bytes.Buffer

	if t, found := cache.Get(path); found {
		var tb bytes.Buffer
		if err := t.(*template.Template).ExecuteTemplate(&tb, path, model); err != nil {
			return "", errors.WithStack(err)
		}
		return tb.String(), nil
	}

	if file, err := box.Open(path); err == nil {
		defer file.Close()
		if _, err := io.Copy(&b, file); err != nil {
			return "", errors.WithStack(err)
		}
	} else {
		file, err := os.Open(path)
		if err != nil {
			return "", errors.WithStack(err)
		}
		defer file.Close()
		if _, err := io.Copy(&b, file); err != nil {
			return "", errors.WithStack(err)
		}
	}

	t, err := template.New(path).Funcs(sprig.TxtFuncMap()).Parse(b.String())
	if err != nil {
		return "", errors.WithStack(err)
	}

	_ = cache.Add(path, t)
	var tb bytes.Buffer
	if err := t.ExecuteTemplate(&tb, path, model); err != nil {
		return "", errors.WithStack(err)
	}

	return tb.String(), nil
}
