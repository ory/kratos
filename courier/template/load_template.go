// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"bytes"
	"context"
	"embed"
	htemplate "html/template"
	"io"
	"io/fs"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/x/fetcher"
	"github.com/ory/x/httpx"

	"github.com/Masterminds/sprig/v3"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

//go:embed courier/builtin/templates/*
var templates embed.FS

var Cache, _ = lru.New(16)

type Template interface {
	Execute(wr io.Writer, data interface{}) error
}

type templateDependencies interface {
	HTTPClient(ctx context.Context, opts ...httpx.ResilientOptions) *retryablehttp.Client
}

func loadBuiltInTemplate(filesystem fs.FS, name string, html bool) (Template, error) {
	if t, found := Cache.Get(name); found {
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

	_ = Cache.Add(name, tpl)
	return tpl, nil
}

func loadRemoteTemplate(ctx context.Context, d templateDependencies, url string, html bool) (Template, error) {
	var b []byte
	var err error

	// instead of creating a new request always we always cache the bytes.Buffer using the url as the key
	if t, found := Cache.Get(url); found {
		b = t.([]byte)
	} else {
		f := fetcher.NewFetcher(fetcher.WithClient(d.HTTPClient(ctx)))
		bb, err := f.FetchContext(ctx, url)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		b = bb.Bytes()
		_ = Cache.Add(url, b)
	}

	var t Template
	if html {
		t, err = htemplate.New(url).Funcs(sprig.HermeticHtmlFuncMap()).Parse(string(b))
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		t, err = template.New(url).Funcs(sprig.HermeticTxtFuncMap()).Parse(string(b))
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return t, nil
}

func loadTemplate(filesystem fs.FS, name, pattern string, html bool) (Template, error) {
	if t, found := Cache.Get(name); found {
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
		t, err := htemplate.New(filepath.Base(name)).Funcs(sprig.HermeticHtmlFuncMap()).ParseFS(filesystem, glob)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = t
	} else {
		t, err := template.New(filepath.Base(name)).Funcs(sprig.HermeticTxtFuncMap()).ParseFS(filesystem, glob)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tpl = t
	}

	_ = Cache.Add(name, tpl)
	return tpl, nil
}

func LoadText(ctx context.Context, d templateDependencies, filesystem fs.FS, name, pattern string, model interface{}, remoteURL string) (string, error) {
	var t Template
	var err error
	if remoteURL != "" {
		t, err = loadRemoteTemplate(ctx, d, remoteURL, false)
		if err != nil {
			return "", err
		}
	} else {
		t, err = loadTemplate(filesystem, name, pattern, false)
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

func LoadHTML(ctx context.Context, d templateDependencies, filesystem fs.FS, name, pattern string, model interface{}, remoteURL string) (string, error) {
	var t Template
	var err error
	if remoteURL != "" {
		t, err = loadRemoteTemplate(ctx, d, remoteURL, true)
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
