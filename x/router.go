package x

import (
	"net/http"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

type RouterPublic struct {
	*httprouter.Router
}

func NewRouterPublic() *RouterPublic {
	return &RouterPublic{
		Router: httprouter.New(),
	}
}

func (r *RouterPublic) GET(path string, handle httprouter.Handle) {
	r.Handle("GET", path, NoCacheHandle(handle))
}

func (r *RouterPublic) HEAD(path string, handle httprouter.Handle) {
	r.Handle("HEAD", path, NoCacheHandle(handle))
}

func (r *RouterPublic) POST(path string, handle httprouter.Handle) {
	r.Handle("POST", path, NoCacheHandle(handle))
}

func (r *RouterPublic) PUT(path string, handle httprouter.Handle) {
	r.Handle("PUT", path, NoCacheHandle(handle))
}

func (r *RouterPublic) PATCH(path string, handle httprouter.Handle) {
	r.Handle("PATCH", path, NoCacheHandle(handle))
}

func (r *RouterPublic) DELETE(path string, handle httprouter.Handle) {
	r.Handle("DELETE", path, NoCacheHandle(handle))
}

func (r *RouterPublic) Handle(method, path string, handle httprouter.Handle) {
	r.Router.Handle(method, path, NoCacheHandle(handle))
}

func (r *RouterPublic) HandlerFunc(method, path string, handler http.HandlerFunc) {
	r.Router.HandlerFunc(method, path, NoCacheHandlerFunc(handler))
}

func (r *RouterPublic) Handler(method, path string, handler http.Handler) {
	r.Router.Handler(method, path, NoCacheHandler(handler))
}

type RouterAdmin struct {
	*httprouter.Router
}

func NewRouterAdmin() *RouterAdmin {
	return &RouterAdmin{
		Router: httprouter.New(),
	}
}

func (r *RouterAdmin) GET(path string, handle httprouter.Handle) {
	r.Router.GET(filepath.Join(AdminPrefix, path), NoCacheHandle(handle))
}

func (r *RouterAdmin) HEAD(path string, handle httprouter.Handle) {
	r.Router.HEAD(filepath.Join(AdminPrefix, path), NoCacheHandle(handle))
}

func (r *RouterAdmin) POST(path string, handle httprouter.Handle) {
	r.Router.POST(filepath.Join(AdminPrefix, path), NoCacheHandle(handle))
}

func (r *RouterAdmin) PUT(path string, handle httprouter.Handle) {
	r.Router.PUT(filepath.Join(AdminPrefix, path), NoCacheHandle(handle))
}

func (r *RouterAdmin) PATCH(path string, handle httprouter.Handle) {
	r.Router.PATCH(filepath.Join(AdminPrefix, path), NoCacheHandle(handle))
}

func (r *RouterAdmin) DELETE(path string, handle httprouter.Handle) {
	r.Router.DELETE(filepath.Join(AdminPrefix, path), NoCacheHandle(handle))
}

func (r *RouterAdmin) Handle(method, path string, handle httprouter.Handle) {
	r.Router.Handle(method, filepath.Join(AdminPrefix, path), NoCacheHandle(handle))
}

func (r *RouterAdmin) HandlerFunc(method, path string, handler http.HandlerFunc) {
	r.Router.HandlerFunc(method, filepath.Join(AdminPrefix, path), NoCacheHandlerFunc(handler))
}

func (r *RouterAdmin) Handler(method, path string, handler http.Handler) {
	r.Router.Handler(method, filepath.Join(AdminPrefix, path), NoCacheHandler(handler))
}

func (r *RouterAdmin) Lookup(method, path string) {
	r.Router.Lookup(method, filepath.Join(AdminPrefix, path))
}
