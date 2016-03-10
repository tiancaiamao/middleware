package debug

import (
	"net/http"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
	Debugln(args ...interface{})
}

type debug struct {
	Logger
}

func (this debug) Chain(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		this.Debugln(r.Method, r.RequestURI)
		h.ServeHTTP(w, r)
	})
}

func New(log Logger) debug {
	return debug{
		Logger: log,
	}
}
