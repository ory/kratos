package template

import (
	"bytes"
	"embed"
	"io"
	"os"
	"text/template"

	_ "embed"

	"github.com/Masterminds/sprig/v3"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

//go:embed courier/builtin/templates/*
var templates embed.FS

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

	if file, err := templates.Open(path); err == nil {
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
