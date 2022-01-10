package template

import (
	"bytes"
	"embed"
	htemplate "html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

//go:embed courier/builtin/templates/*
var templates embed.FS

var cache, _ = lru.New(16)

type Template interface {
	Execute(wr io.Writer, data interface{}) error
}

func loadBuiltInTemplate(osdir, name string, html bool) (Template, error) {
	if t, found := cache.Get(name); found {
		return t.(Template), nil
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

	var tpl Template
	if html {
		t, err := htemplate.New(name).Funcs(sprig.HtmlFuncMap()).Parse(b.String())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = t
	} else {
		t, err := template.New(name).Funcs(sprig.TxtFuncMap()).Parse(b.String())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = t
	}

	_ = cache.Add(name, tpl)
	return tpl, nil
}

func loadTemplate(osdir, name, pattern string, html bool) (Template, error) {
	if t, found := cache.Get(name); found {
		return t.(Template), nil
	}

	// make sure osdir and template name exists, otherwise fallback to built in templates
	f, _ := filepath.Glob(path.Join(osdir, name))
	if f == nil {
		return loadBuiltInTemplate(osdir, name, html)
	}

	// if pattern is defined, use it for glob
	var glob string = name
	if pattern != "" {
		m, _ := filepath.Glob(path.Join(osdir, pattern))
		if m != nil {
			glob = pattern
		}
	}

	var tpl Template
	if html {
		t, err := htemplate.New(filepath.Base(name)).Funcs(sprig.HtmlFuncMap()).ParseGlob(path.Join(osdir, glob))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = t
	} else {
		t, err := template.New(filepath.Base(name)).Funcs(sprig.TxtFuncMap()).ParseGlob(path.Join(osdir, glob))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = t
	}

	_ = cache.Add(name, tpl)
	return tpl, nil
}

func LoadTextTemplate(osdir, name, pattern string, model interface{}) (string, error) {
	t, err := loadTemplate(osdir, name, pattern, false)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, model); err != nil {
		return "", err
	}
	return b.String(), nil
}

func LoadHTMLTemplate(osdir, name, pattern string, model interface{}) (string, error) {
	t, err := loadTemplate(osdir, name, pattern, true)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, model); err != nil {
		return "", err
	}
	return b.String(), nil
}
