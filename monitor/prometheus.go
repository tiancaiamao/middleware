package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tiancaiamao/middleware"
	"net/http"
)

func New(handlerName string) middleware.MiddleWare {
	return middleware.MiddleWareFunc(func(h http.Handler) http.Handler {
		return prometheus.InstrumentHandler(handlerName, h)
	})
}
