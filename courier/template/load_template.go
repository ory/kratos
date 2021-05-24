package template

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"os"
	"path"
	"text/template"

	_ "embed"

	"github.com/Masterminds/sprig/v3"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

//go:embed courier/builtin/templates/*
var templates embed.FS

func openTemplate(dir, name string) (fs.File, error) {
	if file, err := os.DirFS(dir).Open(name); err == nil {
		return file, nil
	}
	// hard-code the dir, since this is embedded at build time (see above)
	return templates.Open(path.Join("courier/builtin/templates", name))
}

var cache, _ = lru.New(16)

func loadTextTemplate(dir, name string, model interface{}) (string, error) {
	var b bytes.Buffer

	if t, found := cache.Get(name); found {
		var tb bytes.Buffer
		if err := t.(*template.Template).ExecuteTemplate(&tb, name, model); err != nil {
			return "", errors.WithStack(err)
		}
		return tb.String(), nil
	}

	file, err := openTemplate(dir, name)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer file.Close()
	if _, err := io.Copy(&b, file); err != nil {
		return "", errors.WithStack(err)
	}

	t, err := template.New(name).Funcs(sprig.TxtFuncMap()).Parse(b.String())
	if err != nil {
		return "", errors.WithStack(err)
	}

	_ = cache.Add(name, t)
	var tb bytes.Buffer
	if err := t.ExecuteTemplate(&tb, name, model); err != nil {
		return "", errors.WithStack(err)
	}

	return tb.String(), nil
}
