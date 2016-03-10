API in [godoc](http://godoc.org/github.com/tiancaiamao/middleware)

The trick played here is that context is hiden in ResponseWriter. So, one can write standard http.Handler as usual
and get the context as need, in this way:

	func HelloWorld(w http.ResponseWriter, r *http.Request) {
		if cw, ok := w.(middleware.ContextResponseWriter); ok {
			value := cw.Value("key")
			...
		}
	}

As elegancy as gorilla while global lock is unnecessary. 

A complete demo:

	package main
	
	import (
		"github.com/tiancaiamao/middleware"
		"io"
		"net/http"
	)
	
	type MyContextResponseWriter struct {
		http.ResponseWriter
		key, value interface{}
	}
	
	func (w *MyContextResponseWriter) Value(key interface{}) interface{} {
		if w.key == key {
			return w.value
		}
		return nil
	}
	
	func HelloWorld(w http.ResponseWriter, r *http.Request) {
		if ctx, ok := w.(middleware.ContextResponseWriter); ok {
			valueFromContext := ctx.Value("xxx")
			io.WriteString(w, "hello, "+valueFromContext.(string))
			return
		}
		
		io.WriteString(w, "hello world")
	}
	
	type MiddleWareDemo struct{}
	
	func (demo MiddleWareDemo) Chain(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cw := &MyContextResponseWriter{
				ResponseWriter: w,
				key:            "xxx",
				value:          "demo",
			}
			
			h.ServeHTTP(cw, r)
		})
	}
	
	func main() {
		handler := middleware.New(http.HandlerFunc(HelloWorld), MiddleWareDemo{})
		http.ListenAndServe(":8080", handler)
	}

Tips: 

1. golang/x/net/context can be used as ContextResponseWriter's Value if you want. 
2. the code is deadly simple, so you may copy one in your project rather then import this package.
