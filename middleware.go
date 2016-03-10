package middleware

import (
	"net/http"
)

// MiddleWare is a interface that accept a http.Handler and return a http.Handler
type MiddleWare interface {
	Chain(http.Handler) http.Handler
}

// MiddleWareFunc is to MiddleWare what http.HandlerFunc is to http.Handler
type MiddleWareFunc func(http.Handler) http.Handler

// MiddleWareFunc implement the MiddleWare interface
func (f MiddleWareFunc) Chain(h http.Handler) http.Handler {
	return f(h)
}

type Context interface {
	Value(interface{}) interface{}
}

// ContextResponseWriter is a interface that require Value method in addition to http.ResponseWriter
//
// golang/x/net/context.Context can be used as Value here because it implements Value(interface{})interface{}
type ContextResponseWriter interface {
	http.ResponseWriter
	Context
}

type Chain struct {
	middlewares []MiddleWare
	raw         http.Handler
}

// New(Raw, A, B, C) return a *Chain which, when called, will execute in this order: C -> B -> A -> Raw.
//
// Raw is the original Handler, wrap it with middleware A,
// then wrap the result Handler with middleware B,
// and then wrap the result Handler with middleware C
func New(handler http.Handler, middlewares ...MiddleWare) *Chain {
	return &Chain{
		raw:         handler,
		middlewares: middlewares,
	}
}

func (c *Chain) Add(middleware MiddleWare) {
	c.middlewares = append(c.middlewares, middleware)
}

// Chain implement the http.Handler interface
func (c *Chain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	final := c.raw
	for _, mw := range c.middlewares {
		final = mw.Chain(final)
	}
	final.ServeHTTP(w, r)
}
