package log

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

// TODO this is kinda hacked together, give it a once-over check

// NewHttpLogger returns an HttpLogger, suitable for use as http middleware.
func NewHttpLogger(service string) func(h http.Handler) http.Handler {
	middleware := func(h http.Handler) http.Handler {
		entry := logrus.WithFields(logrus.Fields{
			"service": service,
		})

		fn := func(w http.ResponseWriter, r *http.Request) {
			lw := &basicWriter{ResponseWriter: w}

			entry.WithFields(logrus.Fields{
				"uri":    r.URL.String(),
				"method": r.Method,
				"remote": r.RemoteAddr,
			}).Info("Beginning request processing")

			t1 := time.Now()
			h.ServeHTTP(lw, r)

			lw.maybeWriteHeader()

			entry.WithFields(logrus.Fields{
				"status": lw.status(),
				"uri":    r.URL.String(),
				"method": r.Method,
				"remote": r.RemoteAddr,
				"wall":   time.Now().Sub(t1).String(),
			}).Info("Request processing complete")
		}

		return http.HandlerFunc(fn)
	}

	return middleware
}

// basicWriter wraps a http.ResponseWriter that implements the minimal
// http.ResponseWriter interface.
type basicWriter struct {
	http.ResponseWriter
	wroteHeader bool
	code        int
	bytes       int
}

func (b *basicWriter) WriteHeader(code int) {
	if !b.wroteHeader {
		b.code = code
		b.wroteHeader = true
		b.ResponseWriter.WriteHeader(code)
	}
}

func (b *basicWriter) Write(buf []byte) (int, error) {
	b.maybeWriteHeader()
	return b.ResponseWriter.Write(buf)
}

func (b *basicWriter) maybeWriteHeader() {
	if !b.wroteHeader {
		b.WriteHeader(http.StatusOK)
	}
}

func (b *basicWriter) status() int {
	return b.code
}

func (b *basicWriter) Unwrap() http.ResponseWriter {
	return b.ResponseWriter
}
