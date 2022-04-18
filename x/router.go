package x

import (
	"net/http"
	"path"

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

func (r *RouterAdmin) GET(publicPath string, handle httprouter.Handle) {
	r.Router.GET(path.Join(AdminPrefix, publicPath), NoCacheHandle(handle))
}

func (r *RouterAdmin) HEAD(publicPath string, handle httprouter.Handle) {
	r.Router.HEAD(path.Join(AdminPrefix, publicPath), NoCacheHandle(handle))
}

func (r *RouterAdmin) POST(publicPath string, handle httprouter.Handle) {
	r.Router.POST(path.Join(AdminPrefix, publicPath), NoCacheHandle(handle))
}

func (r *RouterAdmin) PUT(publicPath string, handle httprouter.Handle) {
	r.Router.PUT(path.Join(AdminPrefix, publicPath), NoCacheHandle(handle))
}

func (r *RouterAdmin) PATCH(publicPath string, handle httprouter.Handle) {
	r.Router.PATCH(path.Join(AdminPrefix, publicPath), NoCacheHandle(handle))
}

func (r *RouterAdmin) DELETE(publicPath string, handle httprouter.Handle) {
	r.Router.DELETE(path.Join(AdminPrefix, publicPath), NoCacheHandle(handle))
}

func (r *RouterAdmin) Handle(method, publicPath string, handle httprouter.Handle) {
	r.Router.Handle(method, path.Join(AdminPrefix, publicPath), NoCacheHandle(handle))
}

func (r *RouterAdmin) HandlerFunc(method, publicPath string, handler http.HandlerFunc) {
	r.Router.HandlerFunc(method, path.Join(AdminPrefix, publicPath), NoCacheHandlerFunc(handler))
}

func (r *RouterAdmin) Handler(method, publicPath string, handler http.Handler) {
	r.Router.Handler(method, path.Join(AdminPrefix, publicPath), NoCacheHandler(handler))
}

func (r *RouterAdmin) Lookup(method, publicPath string) {
	r.Router.Lookup(method, path.Join(AdminPrefix, publicPath))
}
