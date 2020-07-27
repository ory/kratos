package x

import "github.com/julienschmidt/httprouter"

type RouterAdmin struct {
	*httprouter.Router
}

type RouterPublic struct {
	*httprouter.Router
}

func NewRouterPublic() *RouterPublic {
	return &RouterPublic{
		Router: httprouter.New(),
	}
}

func NewRouterAdmin() *RouterAdmin {
	return &RouterAdmin{
		Router: httprouter.New(),
	}
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *RouterPublic) GET(path string, handle httprouter.Handle) {
	r.Handle("GET", path, NoCacheHandler(handle))
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *RouterPublic) HEAD(path string, handle httprouter.Handle) {
	r.Handle("HEAD", path, NoCacheHandler(handle))
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *RouterPublic) POST(path string, handle httprouter.Handle) {
	r.Handle("POST", path, NoCacheHandler(handle))
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *RouterPublic) PUT(path string, handle httprouter.Handle) {
	r.Handle("PUT", path, NoCacheHandler(handle))
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *RouterPublic) PATCH(path string, handle httprouter.Handle) {
	r.Handle("PATCH", path, NoCacheHandler(handle))
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *RouterPublic) DELETE(path string, handle httprouter.Handle) {
	r.Handle("DELETE", path, NoCacheHandler(handle))
}
