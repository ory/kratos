package template

import (
	"bytes"
	"embed"
	"io"
	"os"
	"path"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

//go:embed courier/builtin/templates/*
var templates embed.FS

var cache, _ = lru.New(16)

func loadTemplate(osdir, name string) (*template.Template, error) {
	if t, found := cache.Get(name); found {
		return t.(*template.Template), nil
	}

	file, err := os.DirFS(osdir).Open(name)
	if err != nil {
		// try to fallback to bundled templates
		var fallbackErr error
		file, fallbackErr = templates.Open(path.Join("courier/builtin/templates", name))
		if fallbackErr != nil {
			// return original error from os.DirFS
			return nil, errors.WithStack(err)
		}
	}

	defer file.Close()

	var b bytes.Buffer
	if _, err := io.Copy(&b, file); err != nil {
		return nil, errors.WithStack(err)
	}

	t, err := template.New(name).Funcs(sprig.TxtFuncMap()).Parse(b.String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	_ = cache.Add(name, t)
	return t, nil
}

func loadTextTemplate(osdir, name string, model interface{}) (string, error) {
	t, err := loadTemplate(osdir, name)
	if err != nil {
		return "", err
	}
	var tb bytes.Buffer
	if err := t.ExecuteTemplate(&tb, name, model); err != nil {
		return "", errors.WithStack(err)
	}
	return tb.String(), nil
}
