package trace

import (
	"net/http"
	"strings"
	"time"

	"github.com/tiancaiamao/middleware"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/appdash/httptrace"
)

type responseInfoRecorder struct {
	http.ResponseWriter
	appdash.SpanID

	statusCode    int   // HTTP response status code
	ContentLength int64 // number of bytes written using the Write method
}

// Write always succeeds and writes to r.Body, if not nil.
func (r *responseInfoRecorder) Write(b []byte) (int, error) {
	r.ContentLength += int64(len(b))
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

func (r *responseInfoRecorder) StatusCode() int {
	if r.statusCode == 0 {
		return http.StatusOK
	}
	return r.statusCode
}

// WriteHeader sets r.Code.
func (r *responseInfoRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// partialResponse constructs a partial response object based on the
// information it is able to determine about the response.
func (r *responseInfoRecorder) partialResponse() *http.Response {
	return &http.Response{
		StatusCode:    r.StatusCode(),
		ContentLength: r.ContentLength,
		Header:        r.Header(),
	}
}

// Flush implements the http.Flusher interface and sends any buffered
// data to the client, if the underlying http.ResponseWriter itself
// implements http.Flusher.
func (r *responseInfoRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *responseInfoRecorder) Value(interface{}) interface{} {
	return r.SpanID
}

func New(c appdash.Collector) middleware.MiddleWare {
	return middleware.MiddleWareFunc(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			spanID, err := httptrace.GetSpanID(r.Header)
			if err != nil {
				*spanID = appdash.NewRootSpanID()
			}

			e := httptrace.NewServerEvent(r)
			e.ServerRecv = time.Now()

			rr := &responseInfoRecorder{
				ResponseWriter: w,
				SpanID:         *spanID,
			}

			h.ServeHTTP(rr, r)

			httptrace.SetSpanIDHeader(rr.Header(), *spanID)
			e.Route = r.URL.Path
			e.Response = responseInfo(rr.partialResponse())
			e.ServerSend = time.Now()

			rec := appdash.NewRecorder(*spanID, c)
			rec.Name(e.Route)
			rec.Event(e)
		})
	})
}

func responseInfo(r *http.Response) httptrace.ResponseInfo {
	return httptrace.ResponseInfo{
		Headers:       redactHeaders(r.Header, r.Trailer),
		ContentLength: r.ContentLength,
		StatusCode:    r.StatusCode,
	}
}

var (
	redacted = []string{"REDACTED"}
)

func redactHeaders(header, trailer http.Header) map[string]string {
	h := make(http.Header, len(header)+len(trailer))
	for k, v := range header {
		if isRedacted(k) {
			h[k] = redacted
		} else {
			h[k] = v
		}
	}
	for k, v := range trailer {
		if isRedacted(k) {
			h[k] = redacted
		} else {
			h[k] = append(h[k], v...)
		}
	}
	m := make(map[string]string, len(h))
	for k, v := range h {
		m[http.CanonicalHeaderKey(k)] = strings.Join(v, ",")
	}
	return m
}

func isRedacted(name string) bool {
	for _, v := range httptrace.RedactedHeaders {
		if strings.EqualFold(name, v) {
			return true
		}
	}
	return false
}
