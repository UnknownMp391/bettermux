package bettermux

import (
	"net/http"
	"strings"
)

// BetterMux is a group routing and middleware wrapper around the standard
// library's http.ServeMux. It has zero runtime overhead: all of its logic
// runs at route registration time, and serving is delegated entirely to the
// underlying *http.ServeMux.
//
// The API is inspired by github.com/go-chi/chi.
type BetterMux struct {
	// RealMux is the underlying *http.ServeMux that all routes are
	// registered on and that ServeHTTP delegates to.
	RealMux *http.ServeMux

	// ParentMux points to the BetterMux this one was derived from via With,
	// and is used to walk the middleware chain at route registration time.
	// It is nil for a mux created by NewBetterMux or
	// NewBetterMuxWithServeMux.
	ParentMux *BetterMux

	// Middleware is the middleware function declared by With on this mux.
	// It is nil unless the mux was created by With.
	Middleware func(http.Handler) http.Handler
}

// NewBetterMux returns a new BetterMux backed by a fresh http.ServeMux.
func NewBetterMux() *BetterMux {
	return &BetterMux{
		RealMux: http.NewServeMux(),
	}
}

// NewBetterMuxWithServeMux returns a new BetterMux that wraps the given
// *http.ServeMux. If mux is nil, a new http.ServeMux is created.
func NewBetterMuxWithServeMux(mux *http.ServeMux) *BetterMux {
	if mux == nil {
		mux = http.NewServeMux()
	}
	return &BetterMux{
		RealMux: mux,
	}
}

// ServeHTTP implements http.Handler by delegating to the underlying
// http.ServeMux.
func (mux *BetterMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.RealMux.ServeHTTP(w, r)
}

// Handle registers handler for the given pattern on the underlying
// http.ServeMux, applying the mux's middleware chain (declared via With)
// to the handler first.
func (mux *BetterMux) Handle(pattern string, handler http.Handler) {
	mux.RealMux.Handle(pattern, mux.applyMiddlewareChain(handler))
}

// HandleFunc registers the handler function fn for the given pattern,
// equivalent to Handle(pattern, fn).
func (mux *BetterMux) HandleFunc(pattern string, fn http.HandlerFunc) {
	mux.Handle(pattern, fn)
}

// Get registers handler for the "GET <pattern>" pattern, using the
// Go 1.22+ enhanced ServeMux routing (pattern may contain wildcards such
// as "/users/{id}", read via r.PathValue).
func (mux *BetterMux) Get(pattern string, handler http.HandlerFunc) {
	mux.Handle("GET "+pattern, handler)
}

// Post registers handler for the "POST <pattern>" pattern, using the
// Go 1.22+ enhanced ServeMux routing.
func (mux *BetterMux) Post(pattern string, handler http.HandlerFunc) {
	mux.Handle("POST "+pattern, handler)
}

// Put registers handler for the "PUT <pattern>" pattern, using the
// Go 1.22+ enhanced ServeMux routing.
func (mux *BetterMux) Put(pattern string, handler http.HandlerFunc) {
	mux.Handle("PUT "+pattern, handler)
}

// Delete registers handler for the "DELETE <pattern>" pattern, using the
// Go 1.22+ enhanced ServeMux routing.
func (mux *BetterMux) Delete(pattern string, handler http.HandlerFunc) {
	mux.Handle("DELETE "+pattern, handler)
}

// Patch registers handler for the "PATCH <pattern>" pattern, using the
// Go 1.22+ enhanced ServeMux routing.
func (mux *BetterMux) Patch(pattern string, handler http.HandlerFunc) {
	mux.Handle("PATCH "+pattern, handler)
}

// Options registers handler for the "OPTIONS <pattern>" pattern, using
// the Go 1.22+ enhanced ServeMux routing.
func (mux *BetterMux) Options(pattern string, handler http.HandlerFunc) {
	mux.Handle("OPTIONS "+pattern, handler)
}

// Mount attaches handler to the given path prefix. The prefix is stripped
// from the request path before it reaches handler (via http.StripPrefix),
// so a handler mounted at "/static" receives requests rooted at "/".
//
// Both "/static" and "/static/" are accepted and produce the same
// registration (the subtree pattern "/static/").
func (mux *BetterMux) Mount(path string, handler http.Handler) {
	var stripPath string
	var routePath string

	if before, ok := strings.CutSuffix(path, "/"); ok {
		stripPath = before
		routePath = path
	} else {
		stripPath = path
		routePath = path + "/"
	}

	mux.Handle(routePath, http.StripPrefix(stripPath, handler))
}

// Route creates a child BetterMux, runs fn on it to register routes, and
// mounts the child at the given path prefix (see Mount). It returns the
// child mux so routes can also be registered afterwards.
func (mux *BetterMux) Route(path string, fn func(*BetterMux)) *BetterMux {

	childRealMux := http.NewServeMux()
	childMux := &BetterMux{
		RealMux: childRealMux,
	}

	if fn != nil {
		fn(childMux)
	}

	mux.Mount(path, childMux)

	return childMux
}

// With returns a new BetterMux that shares the same underlying
// http.ServeMux and declares mfn as middleware for it. With can be
// chained; middleware declared first is applied first (outermost). The
// middleware chain is applied to a handler when it is registered via
// Handle, HandleFunc, or any of the method helpers such as Get.
func (mux *BetterMux) With(mfn func(http.Handler) http.Handler) *BetterMux {
	return &BetterMux{
		RealMux:    mux.RealMux,
		ParentMux:  mux,
		Middleware: mfn,
	}
}

func (mux *BetterMux) applyMiddlewareChain(handler http.Handler) http.Handler {
	h := handler
	cur := mux
	for cur != nil {
		if cur.Middleware != nil {
			h = cur.Middleware(h)
		}
		cur = cur.ParentMux
	}
	return h
}
