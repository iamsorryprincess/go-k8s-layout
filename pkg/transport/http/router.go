package http

import (
	"net/http"
	"slices"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) Use(middleware func(http.Handler) http.Handler) *Router {
	handler := middleware(r.mux)

	mux, ok := handler.(*http.ServeMux)
	if !ok {
		panic("unexpected behavior: http.Handler is not http.Mux")
	}

	r.mux = mux
	return r
}

func (r *Router) Handle(pattern string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) *Router {
	if len(middlewares) > 0 {
		handler = withHandler(handler, middlewares...)
	}

	r.mux.Handle(pattern, handler)

	return r
}

func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request), middlewares ...func(http.Handler) http.Handler) *Router {
	if len(middlewares) > 0 {
		newHandler := withHandlerFunc(handler, middlewares...)
		r.mux.Handle(pattern, newHandler)
		return r
	}

	r.mux.HandleFunc(pattern, handler)

	return r
}

func withHandler(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range slices.Backward(middlewares) {
		handler = middleware(handler)
	}
	return handler
}

func withHandlerFunc(handler http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return withHandler(handler, middlewares...)
}
