package template

import (
	"bytes"
	"embed"
	htemplate "html/template"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

//go:embed courier/builtin/templates/*
var templates embed.FS

var cache, _ = lru.New(16)

func loadBuiltInTemplate(osdir, name string, html bool) (interface{}, error) {
	if t, found := cache.Get(name); found {
		return t, nil
	}

	file, err := os.DirFS(osdir).Open(name)
	if err != nil {
		// try to fallback to bundled templates
		var fallbackErr error
		file, fallbackErr = templates.Open(filepath.Join("courier/builtin/templates", name))
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

	var tpl interface{}
	if html {
		tpl, err = htemplate.New(name).Funcs(sprig.HtmlFuncMap()).Parse(b.String())
	} else {
		tpl, err = template.New(name).Funcs(sprig.TxtFuncMap()).Parse(b.String())
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	_ = cache.Add(name, tpl)
	return tpl, nil
}

func loadTemplate(osdir, name, pattern string, html bool) (interface{}, error) {
	if t, found := cache.Get(name); found {
		return t, nil
	}

	// make sure osdir and template name exists, otherwise fallback to built in templates
	f, _ := filepath.Glob(filepath.Join(osdir, name))
	if f == nil {
		return loadBuiltInTemplate(osdir, name, html)
	}

	// if pattern is defined, use it for glob
	var glob string = name
	if pattern != "" {
		m, _ := filepath.Glob(filepath.Join(osdir, pattern))
		if m != nil {
			glob = pattern
		}
	}

	// parse templates matching glob
	var tpl interface{}
	var err error
	if html {
		tpl, _ = htemplate.New(filepath.Base(name)).Funcs(sprig.HtmlFuncMap()).ParseGlob(filepath.Join(osdir, glob))
	} else {
		tpl, _ = template.New(filepath.Base(name)).Funcs(sprig.TxtFuncMap()).ParseGlob(filepath.Join(osdir, glob))
	}

	if err != nil || tpl == nil {
		return nil, errors.WithStack(err)
	}

	_ = cache.Add(name, tpl)
	return tpl, nil
}

func loadTextTemplate(osdir, name, pattern string, model interface{}) (string, error) {
	t, err := loadTemplate(osdir, name, pattern, false)
	if err != nil {
		return "", err
	}

	var tb bytes.Buffer
	if err := t.(*template.Template).Execute(&tb, model); err != nil {
		return "", errors.WithStack(err)
	}

	return tb.String(), nil
}

func loadHTMLTemplate(osdir, name, pattern string, model interface{}) (string, error) {
	t, err := loadTemplate(osdir, name, pattern, true)
	if err != nil {
		return "", err
	}

	var tb bytes.Buffer
	if err := t.(*htemplate.Template).Execute(&tb, model); err != nil {
		return "", errors.WithStack(err)
	}

	return tb.String(), nil
}
