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

type ExecutableTemplate struct {
	Template interface {
		Execute(wr io.Writer, data interface{}) error
	}
}

func newExecutableHTMLTemplate(t *htemplate.Template) *ExecutableTemplate {
	return &ExecutableTemplate{Template: t}
}

func newExecutableTextTemplate(t *template.Template) *ExecutableTemplate {
	return &ExecutableTemplate{Template: t}
}

func (t *ExecutableTemplate) Execute(data interface{}) (string, error) {
	var tb bytes.Buffer
	if err := t.Template.Execute(&tb, data); err != nil {
		return "", errors.WithStack(err)
	}

	return tb.String(), nil
}

func loadBuiltInTemplate(osdir, name string, html bool) (*ExecutableTemplate, error) {
	if t, found := cache.Get(name); found {
		return t.(*ExecutableTemplate), nil
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

	var tpl *ExecutableTemplate
	if html {
		t, err := htemplate.New(name).Funcs(sprig.HtmlFuncMap()).Parse(b.String())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = newExecutableHTMLTemplate(t)
	} else {
		t, err := template.New(name).Funcs(sprig.TxtFuncMap()).Parse(b.String())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = newExecutableTextTemplate(t)
	}

	_ = cache.Add(name, tpl)
	return tpl, nil
}

func loadTemplate(osdir, name, pattern string, html bool) (*ExecutableTemplate, error) {
	if t, found := cache.Get(name); found {
		return t.(*ExecutableTemplate), nil
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

	var tpl *ExecutableTemplate
	if html {
		t, err := htemplate.New(filepath.Base(name)).Funcs(sprig.HtmlFuncMap()).ParseGlob(path.Join(osdir, glob))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = newExecutableHTMLTemplate(t)
	} else {
		t, err := template.New(filepath.Base(name)).Funcs(sprig.TxtFuncMap()).ParseGlob(path.Join(osdir, glob))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = newExecutableTextTemplate(t)
	}

	_ = cache.Add(name, tpl)
	return tpl, nil
}

func loadTextTemplate(osdir, name, pattern string, model interface{}) (string, error) {
	t, err := loadTemplate(osdir, name, pattern, false)
	if err != nil {
		return "", err
	}
	return t.Execute(model)
}

func loadHTMLTemplate(osdir, name, pattern string, model interface{}) (string, error) {
	t, err := loadTemplate(osdir, name, pattern, true)
	if err != nil {
		return "", err
	}
	return t.Execute(model)
}
