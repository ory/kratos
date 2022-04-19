package request

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/google/go-jsonnet"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/ory/x/fetcher"
	"github.com/ory/x/logrusx"
)

var ErrCancel = errors.New("request cancel by JsonNet")

const (
	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeJSON = "application/json"
)

type Builder struct {
	r           *retryablehttp.Request
	log         *logrusx.Logger
	conf        *Config
	fetchClient *retryablehttp.Client
}

func NewBuilder(config json.RawMessage, client *retryablehttp.Client, l *logrusx.Logger) (*Builder, error) {
	c, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	r, err := retryablehttp.NewRequest(c.Method, c.URL, nil)
	if err != nil {
		return nil, err
	}

	return &Builder{
		r:           r,
		log:         l,
		conf:        c,
		fetchClient: client,
	}, nil
}

func (b *Builder) addAuth() error {
	authConfig := b.conf.Auth

	strategy, err := authStrategy(authConfig.Type, authConfig.Config)
	if err != nil {
		return err
	}

	strategy.apply(b.r)

	return nil
}

func (b *Builder) addBody(body interface{}) error {
	if isNilInterface(body) {
		return nil
	}

	contentType := b.r.Header.Get("Content-Type")

	if b.conf.TemplateURI == "" {
		return errors.New("got empty template path for request with body")
	}

	tpl, err := b.readTemplate()
	if err != nil {
		return err
	}

	switch contentType {
	case ContentTypeForm:
		if err := b.addURLEncodedBody(tpl, body); err != nil {
			return err
		}
	case ContentTypeJSON:
		if err := b.addJSONBody(tpl, body); err != nil {
			return err
		}
	default:
		return errors.New("invalid config - incorrect Content-Type for request with body")
	}

	return nil
}

func (b *Builder) addJSONBody(template *bytes.Buffer, body interface{}) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	if err := enc.Encode(body); err != nil {
		return errors.WithStack(err)
	}

	vm := jsonnet.MakeVM()
	vm.TLACode("ctx", buf.String())

	res, err := vm.EvaluateAnonymousSnippet(b.conf.TemplateURI, template.String())
	if err != nil {
		// Unfortunately we can not use errors.As / errors.Is, see:
		// https://github.com/google/go-jsonnet/issues/592
		if strings.Contains(err.Error(), (&jsonnet.RuntimeError{Msg: "cancel"}).Error()) {
			return errors.WithStack(ErrCancel)
		}

		return errors.WithStack(err)
	}

	rb := strings.NewReader(res)
	b.r.Body = io.NopCloser(rb)
	b.r.ContentLength = int64(rb.Len())

	return nil
}

func (b *Builder) addURLEncodedBody(template *bytes.Buffer, body interface{}) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "")

	if err := enc.Encode(body); err != nil {
		return err
	}

	vm := jsonnet.MakeVM()
	vm.TLACode("ctx", buf.String())

	res, err := vm.EvaluateAnonymousSnippet(b.conf.TemplateURI, template.String())
	if err != nil {
		return err
	}

	values := map[string]string{}
	if err := json.Unmarshal([]byte(res), &values); err != nil {
		return err
	}

	u := url.Values{}

	for key, value := range values {
		u.Add(key, value)
	}

	rb := strings.NewReader(u.Encode())
	b.r.Body = io.NopCloser(rb)

	return nil
}

func (b *Builder) BuildRequest(body interface{}) (*retryablehttp.Request, error) {
	b.r.Header = b.conf.Header
	if err := b.addAuth(); err != nil {
		return nil, err
	}

	// According to the HTTP spec any request method, but TRACE is allowed to
	// have a body. Even this is a bad practice for some of them, like for GET
	if b.conf.Method != http.MethodTrace {
		if err := b.addBody(body); err != nil {
			return nil, err
		}
	}

	return b.r, nil
}

func (b *Builder) readTemplate() (*bytes.Buffer, error) {
	templateURI := b.conf.TemplateURI

	if templateURI == "" {
		return nil, nil
	}

	f := fetcher.NewFetcher(fetcher.WithClient(b.fetchClient))

	tpl, err := f.Fetch(templateURI)
	if errors.Is(err, fetcher.ErrUnknownScheme) {
		// legacy filepath
		templateURI = "file://" + templateURI
		b.log.WithError(err).Warnf("support for filepaths without a 'file://' scheme will be dropped in the next release, please use %s instead in your config", templateURI)

		tpl, err = f.Fetch(templateURI)
	}
	// this handles the first error if it is a known scheme error, or the second fetch error
	if err != nil {
		return nil, err
	}

	return tpl, nil
}

func isNilInterface(i interface{}) bool {
	return i == nil || (reflect.ValueOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil())
}
