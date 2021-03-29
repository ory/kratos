package daemon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"

	"github.com/ory/kratos/driver/config"
)

func TestNewStripPrefixMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		prefixLoader func(r *http.Request) string
		wantNext     bool
		wantPath     string
		wantStatus   int
	}{
		{
			name:         "empty prefix 1",
			target:       "http://127.0.0.1:1234",
			prefixLoader: func(r *http.Request) string { return "" },
			wantNext:     true,
			wantPath:     "",
		},
		{
			name:         "empty prefix 2",
			target:       "http://127.0.0.1:1234/foo",
			prefixLoader: func(r *http.Request) string { return "" },
			wantNext:     true,
			wantPath:     "/foo",
		},
		{
			name:         "slash prefix 1",
			target:       "http://127.0.0.1:1234",
			prefixLoader: func(r *http.Request) string { return "/" },
			wantNext:     true,
			wantPath:     "",
		},
		{
			name:         "slash prefix 2",
			target:       "http://127.0.0.1:1234/foo",
			prefixLoader: func(r *http.Request) string { return "/" },
			wantNext:     true,
			wantPath:     "/foo",
		},
		{
			name:         "slash prefix 3",
			target:       "http://127.0.0.1:1234/",
			prefixLoader: func(r *http.Request) string { return "/" },
			wantNext:     true,
			wantPath:     "/",
		},
		{
			name:         "not matching prefix 1",
			target:       "http://127.0.0.1:1234",
			prefixLoader: func(r *http.Request) string { return "/kratos" },
			wantStatus:   http.StatusNotFound,
			wantPath:     "",
		},
		{
			name:         "not matching prefix 2",
			target:       "http://127.0.0.1:1234/foo",
			prefixLoader: func(r *http.Request) string { return "/kratos" },
			wantStatus:   http.StatusNotFound,
			wantPath:     "/foo",
		},
		{
			name:         "matching prefix 1",
			target:       "http://127.0.0.1:1234/kratos/foo",
			prefixLoader: func(r *http.Request) string { return "/kratos" },
			wantNext:     true,
			wantPath:     "/foo",
		},
		{
			name:         "matching prefix 2",
			target:       "http://127.0.0.1:1234/kratos/foo?query=value",
			prefixLoader: func(r *http.Request) string { return "/kratos" },
			wantNext:     true,
			wantPath:     "/foo",
		},
		{
			name:         "matching prefix 3",
			target:       "http://127.0.0.1:1234/kratos/nested/foo",
			prefixLoader: func(r *http.Request) string { return "/kratos" },
			wantNext:     true,
			wantPath:     "/nested/foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.target, nil)
			mw := NewStripPrefixMiddleware(tt.prefixLoader)
			nextRequest := request
			gotNext := false
			mw(recorder, request, func(w http.ResponseWriter, r *http.Request) {
				gotNext = true
				nextRequest = r
			})
			if tt.wantNext != gotNext {
				t.Errorf("NewStripPrefixMiddleware() executed next %v, want %v", gotNext, tt.wantNext)
			}
			if tt.wantPath != nextRequest.URL.Path {
				t.Errorf("NewStripPrefixMiddleware() got path %s, want %s", nextRequest.URL.Path, tt.wantPath)
			}

			if tt.wantStatus > 0 && recorder.Code != tt.wantStatus {
				t.Errorf("NewStripPrefixMiddleware() got status %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}

func MustURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func Test_extractPrefixFromBaseURL(t *testing.T) {
	tests := []struct {
		name string
		u    *url.URL
		want string
	}{
		{"nil url", nil, ""},
		{"no prefix 1", MustURL("https://127.0.0.1:4433"), ""},
		{"no prefix 2", MustURL("https://127.0.0.1:4433/"), ""},
		{"no prefix 3", MustURL("https://example.com"), ""},
		{"prefix 1", MustURL("https://127.0.0.1:4433/kratos"), "/kratos"},
		{"prefix 2", MustURL("https://example.com/kratos"), "/kratos"},
		{"prefix 3", MustURL("https://example.com/my/kratos"), "/my/kratos"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractPrefixFromBaseURL(tt.u); got != tt.want {
				t.Errorf("extractPrefixFromBaseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockConfigProvider struct {
	Conf *config.Config
}

func (p *MockConfigProvider) Config(ctx context.Context) *config.Config {
	return p.Conf
}

// Ensure that the prefix can be extracted from the configured base url.
func Test_StripPathPrefix(t *testing.T) {
	t.Run("group=public", func(t *testing.T) {
		get := func(cfg *config.Config, target string) string {
			req := httptest.NewRequest(http.MethodGet, target, http.NoBody)
			p := &MockConfigProvider{Conf: cfg}
			return publicURLPrefixExtractor(p)(req)
		}
		c := config.MustNew(logrusx.New("", ""),
			configx.WithConfigFiles("../../internal/.kratos.yaml"))
		c.MustSet(config.ViperKeyPublicBaseURL, "http://public.kratos.ory.sh:1234")
		c.MustSet(config.ViperKeyPublicHost, "public.kratos.ory.sh")
		c.MustSet(config.ViperKeyPublicPort, 1234)
		assert.Equal(t, get(c, "http://public.kratos.ory.sh:1234"), "")

		// no path on base url
		c.MustSet(config.ViperKeyPublicBaseURL, "http://public.kratos.ory.sh:1234")
		assert.Equal(t, get(c, "http://public.kratos.ory.sh:1234"), "")

		// trailing slash on base url
		c.MustSet(config.ViperKeyPublicBaseURL, "http://public.kratos.ory.sh:1234/")
		assert.Equal(t, get(c, "http://public.kratos.ory.sh:1234"), "")

		// base path on base url
		c.MustSet(config.ViperKeyPublicBaseURL, "http://public.kratos.ory.sh:1234/kratos")
		assert.Equal(t, get(c, "http://public.kratos.ory.sh:1234"), "/kratos")

	})

	t.Run("group=admin", func(t *testing.T) {
		get := func(cfg *config.Config, target string) string {
			req := httptest.NewRequest(http.MethodGet, target, http.NoBody)
			p := &MockConfigProvider{Conf: cfg}
			return adminURLPrefixExtractor(p)(req)
		}
		c := config.MustNew(logrusx.New("", ""),
			configx.WithConfigFiles("../../internal/.kratos.yaml"))

		// no base path on url
		c.MustSet(config.ViperKeyAdminBaseURL, "http://admin.kratos.ory.sh:1234")
		assert.Equal(t, get(c, "http://admin.kratos.ory.sh:1234"), "")

		// trailing slash on url
		c.MustSet(config.ViperKeyAdminBaseURL, "http://admin.kratos.ory.sh:1234/")
		assert.Equal(t, get(c, "http://admin.kratos.ory.sh:1234"), "")

		// base path on base url
		c.MustSet(config.ViperKeyAdminBaseURL, "http://admin.kratos.ory.sh:1234/kratos")
		assert.Equal(t, get(c, "http://admin.kratos.ory.sh:1234"), "/kratos")
	})
}
