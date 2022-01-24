package template

import (
	"bytes"
	"embed"
	"github.com/ory/x/fetcher"
	htemplate "html/template"
	"io"
	"io/fs"
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

type options struct {
	root      string
	remoteURL string
}

type LoadTemplateOption func(*options)

func WithRemoteResource(url string) LoadTemplateOption {
	return func(o *options) {
		o.remoteURL = url
	}
}

func WithRemoteResourceRoot(root string) LoadTemplateOption {
	return func(o *options) {
		o.root = root
	}
}

func loadBuiltInTemplate(filesystem fs.FS, name string, html bool) (Template, error) {
	if t, found := cache.Get(name); found {
		return t.(Template), nil
	}

	file, err := filesystem.Open(name)
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

func loadRemoteTemplate(url string, name string, html bool, root string) (Template, error) {
	if t, found := cache.Get(name); found {
		return t.(Template), nil
	}

	f := fetcher.NewFetcher()
	bb, err := f.Fetch(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var t Template
	if html {
		t, err = htemplate.New(root).Funcs(sprig.HtmlFuncMap()).Parse(bb.String())
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		t, err = template.New(root).Funcs(sprig.TxtFuncMap()).Parse(bb.String())
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return t, nil
}

func loadTemplate(filesystem fs.FS, name, pattern string, html bool) (Template, error) {
	if t, found := cache.Get(name); found {
		return t.(Template), nil
	}

	matches, _ := fs.Glob(filesystem, name)
	// make sure the file exists in the fs, otherwise fallback to built in templates
	if matches == nil {
		return loadBuiltInTemplate(filesystem, name, html)
	}

	glob := name
	if pattern != "" {
		// pattern matching is used when we have more than one gotmpl for different use cases, such as i18n support
		// e.g. some_template/template_name* will match some_template/template_name.body.en_US.gotmpl
		matches, _ = fs.Glob(filesystem, pattern)
		// set the glob string to match patterns
		if matches != nil {
			glob = pattern
		}
	}

	var tpl Template
	if html {
		t, err := htemplate.New(filepath.Base(name)).Funcs(sprig.HtmlFuncMap()).ParseFS(filesystem, glob)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = t
	} else {
		t, err := template.New(filepath.Base(name)).Funcs(sprig.TxtFuncMap()).ParseFS(filesystem, glob)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = t
	}

	_ = cache.Add(name, tpl)
	return tpl, nil
}

func LoadTextTemplate(filesystem fs.FS, name, pattern string, model interface{}, remoteURL, remoteTemplateRoot string) (string, error) {
	var t Template
	var err error
	if remoteURL != "" {
		t, err = loadRemoteTemplate(remoteURL, name, false, remoteTemplateRoot)
		if err != nil {
			return "", err
		}
	}

	t, err = loadTemplate(filesystem, name, pattern, false)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	if err := t.Execute(&b, model); err != nil {
		return "", err
	}
	return b.String(), nil
}

func LoadHTMLTemplate(filesystem fs.FS, name, pattern string, model interface{}, remoteURL, remoteTemplateRoot string) (string, error) {
	var t Template
	var err error
	if remoteURL != "" {
		t, err = loadRemoteTemplate(remoteURL, name, true, remoteTemplateRoot)
		if err != nil {
			return "", err
		}
	} else {
		t, err = loadTemplate(filesystem, name, pattern, true)
		if err != nil {
			return "", err
		}
	}

	var b bytes.Buffer
	if err := t.Execute(&b, model); err != nil {
		return "", err
	}
	return b.String(), nil
}
