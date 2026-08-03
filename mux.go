package bettermux

import (
	"net/http"
	"strings"
)

type BetterMux struct {
	realMux    *http.ServeMux
	parentMux  *BetterMux
	middleware func(http.Handler) http.Handler
}

func New() BetterMux {
	return BetterMux{
		realMux: http.NewServeMux(),
	}
}

func NewWithMux(mux *http.ServeMux) BetterMux {
	if mux == nil {
		mux = http.NewServeMux()
	}
	return BetterMux{
		realMux: mux,
	}
}

func (mux BetterMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.realMux.ServeHTTP(w, r)
}

func (mux BetterMux) Handle(pattern string, handler http.Handler) {
	mux.realMux.Handle(pattern, mux.applyMiddlewareChain(handler))
}

func (mux BetterMux) HandleFunc(pattern string, fn http.HandlerFunc) {
	mux.Handle(pattern, fn)
}

func (mux BetterMux) Get(pattern string, handler http.HandlerFunc) {
	mux.Handle("GET "+pattern, handler)
}

func (mux BetterMux) Post(pattern string, handler http.HandlerFunc) {
	mux.Handle("POST "+pattern, handler)
}

func (mux BetterMux) Put(pattern string, handler http.HandlerFunc) {
	mux.Handle("PUT "+pattern, handler)
}

func (mux BetterMux) Delete(pattern string, handler http.HandlerFunc) {
	mux.Handle("DELETE "+pattern, handler)
}

func (mux BetterMux) Patch(pattern string, handler http.HandlerFunc) {
	mux.Handle("PATCH "+pattern, handler)
}

func (mux BetterMux) Options(pattern string, handler http.HandlerFunc) {
	mux.Handle("OPTIONS "+pattern, handler)
}

func (mux BetterMux) Route(path string, fn func(BetterMux)) BetterMux {
	var stripPath string
	var routePath string
	if before, ok := strings.CutSuffix(path, "/"); ok {
		stripPath = before
		routePath = path
	} else {
		stripPath = path
		routePath = path + "/"
	}

	childRealMux := http.NewServeMux()
	childMux := BetterMux{
		realMux: childRealMux,
	}

	if fn != nil {
		fn(childMux)
	}

	mux.Handle(routePath, http.StripPrefix(stripPath, childRealMux))

	return childMux
}

func (mux BetterMux) With(mfn func(http.Handler) http.Handler) BetterMux {
	return BetterMux{
		realMux:    mux.realMux,
		parentMux:  &mux,
		middleware: mfn,
	}
}

func (mux BetterMux) applyMiddlewareChain(handler http.Handler) http.Handler {
	h := handler
	cur := &mux
	for cur != nil {
		if cur.middleware != nil {
			h = cur.middleware(h)
		}
		cur = cur.parentMux
	}
	return h
}
