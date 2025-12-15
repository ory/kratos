// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/ory/x/prometheusx"
)

type RouterPublic struct {
	mux *http.ServeMux
	pmm *prometheusx.MetricsManager
}

type routerDeps interface {
	PrometheusManager() *prometheusx.MetricsManager
}

func NewRouterPublic(deps routerDeps) *RouterPublic {
	return &RouterPublic{
		mux: http.NewServeMux(),
		pmm: deps.PrometheusManager(),
	}
}

// NewTestRouterPublic creates a new RouterPublic for testing purposes without metrics.
func NewTestRouterPublic(*testing.T) *RouterPublic {
	return &RouterPublic{
		mux: http.NewServeMux(),
		pmm: nil, // No metrics manager in test environment
	}
}

func (r *RouterPublic) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *RouterPublic) GET(path string, handler http.HandlerFunc) {
	r.HandlerFunc("GET", path, handler)
}

func (r *RouterPublic) HEAD(path string, handler http.HandlerFunc) {
	r.HandlerFunc("HEAD", path, handler)
}

func (r *RouterPublic) POST(path string, handler http.HandlerFunc) {
	r.HandlerFunc("POST", path, handler)
}

func (r *RouterPublic) PUT(path string, handler http.HandlerFunc) {
	r.HandlerFunc("PUT", path, handler)
}

func (r *RouterPublic) PATCH(path string, handler http.HandlerFunc) {
	r.HandlerFunc("PATCH", path, handler)
}

func (r *RouterPublic) DELETE(path string, handler http.HandlerFunc) {
	r.HandlerFunc("DELETE", path, handler)
}

func (r *RouterPublic) Handle(method, route string, handle http.HandlerFunc) {
	for _, pattern := range []string{
		method + " " + path.Join(route),
		method + " " + path.Join(route, "{$}"),
	} {
		handleWithAllMiddlewares(r.mux, r.pmm, pattern, handle)
	}
}

func (r *RouterPublic) HandlerFunc(method, route string, handler http.HandlerFunc) {
	for _, pattern := range []string{
		method + " " + path.Join(route),
		method + " " + path.Join(route, "{$}"),
	} {
		handleWithAllMiddlewares(r.mux, r.pmm, pattern, handler)
	}
}

func (r *RouterPublic) HandleFunc(pattern string, handler http.HandlerFunc) {
	for _, pattern := range []string{
		path.Join(pattern),
		path.Join(pattern, "{$}"),
	} {
		handleWithAllMiddlewares(r.mux, r.pmm, pattern, handler)
	}
}

func (r *RouterPublic) Handler(method, path string, handler http.Handler) {
	route := method + " " + path
	handleWithAllMiddlewares(r.mux, r.pmm, route, handler)
}

func (r *RouterPublic) HasRoute(method, path string) bool {
	_, pattern := r.mux.Handler(httptest.NewRequest(method, path, nil))
	return pattern != ""
}

type RouterAdmin struct {
	mux *http.ServeMux
	pmm *prometheusx.MetricsManager
}

func NewRouterAdmin(deps routerDeps) *RouterAdmin {
	return &RouterAdmin{
		mux: http.NewServeMux(),
		pmm: deps.PrometheusManager(),
	}
}

// NewTestRouterAdmin creates a new RouterAdmin for testing purposes without metrics.
func NewTestRouterAdmin(*testing.T) *RouterAdmin {
	return &RouterAdmin{
		mux: http.NewServeMux(),
		pmm: nil, // No metrics manager in test environment
	}
}

func (r *RouterAdmin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *RouterAdmin) GET(publicPath string, handler http.HandlerFunc) {
	r.HandlerFunc("GET", publicPath, handler)
}

func (r *RouterAdmin) HEAD(publicPath string, handler http.HandlerFunc) {
	r.HandlerFunc("HEAD", publicPath, handler)
}

func (r *RouterAdmin) POST(publicPath string, handler http.HandlerFunc) {
	r.HandlerFunc("POST", publicPath, handler)
}

func (r *RouterAdmin) PUT(publicPath string, handler http.HandlerFunc) {
	r.HandlerFunc("PUT", publicPath, handler)
}

func (r *RouterAdmin) PATCH(publicPath string, handler http.HandlerFunc) {
	r.HandlerFunc("PATCH", publicPath, handler)
}

func (r *RouterAdmin) DELETE(publicPath string, handler http.HandlerFunc) {
	r.HandlerFunc("DELETE", publicPath, handler)
}

func (r *RouterAdmin) Handle(method, publicPath string, handle http.HandlerFunc) {
	for _, pattern := range []string{
		method + " " + path.Join(AdminPrefix, publicPath),
		method + " " + path.Join(AdminPrefix, publicPath, "{$}"),
	} {
		handleWithAllMiddlewares(r.mux, r.pmm, pattern, http.HandlerFunc(handle))
	}
}

func (r *RouterAdmin) HandlerFunc(method, publicPath string, handler http.HandlerFunc) {
	for _, pattern := range []string{
		method + " " + path.Join(AdminPrefix, publicPath),
		method + " " + path.Join(AdminPrefix, publicPath, "{$}"),
	} {
		handleWithAllMiddlewares(r.mux, r.pmm, pattern, http.HandlerFunc(handler))
	}
}

func (r *RouterAdmin) Handler(method, publicPath string, handler http.Handler) {
	for _, pattern := range []string{
		method + " " + path.Join(AdminPrefix, publicPath),
		method + " " + path.Join(AdminPrefix, publicPath, "{$}"),
	} {
		handleWithAllMiddlewares(r.mux, r.pmm, pattern, (handler))
	}
}

func (r *RouterAdmin) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	for _, p := range []string{
		path.Join(pattern),
		path.Join(pattern, "{$}"),
	} {
		handleWithAllMiddlewares(r.mux, r.pmm, p, http.HandlerFunc(handler))
	}
}

// handleWithAllMiddlewares wraps the handler with NoCache and Prometheus metrics
// middleware if available.
func handleWithAllMiddlewares(mux *http.ServeMux, pmm *prometheusx.MetricsManager, pattern string, handler http.Handler) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		NoCache(w)
		if pmm != nil {
			pmm.ServeHTTP(w, req, func(w http.ResponseWriter, req *http.Request) {
				handler.ServeHTTP(w, req)
			})
		} else {
			handler.ServeHTTP(w, req)
		}
	})
}

type HandlerRegistrar interface {
	RegisterPublicRoutes(public *RouterPublic)
	RegisterAdminRoutes(admin *RouterAdmin)
}
