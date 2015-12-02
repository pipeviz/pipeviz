package log

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/zenazn/goji/web/mutil"
)

// TODO this is kinda hacked together, give it a once-over check

// NewHTTPLogger returns an HTTPLogger, suitable for use as http middleware.
func NewHTTPLogger(system string) func(h http.Handler) http.Handler {
	middleware := func(h http.Handler) http.Handler {
		entry := logrus.WithFields(logrus.Fields{
			"system": system,
		})

		fn := func(w http.ResponseWriter, r *http.Request) {
			lw := mutil.WrapWriter(w)

			entry.WithFields(logrus.Fields{
				"uri":    r.URL.String(),
				"method": r.Method,
				"remote": r.RemoteAddr,
			}).Info("Beginning request processing")

			t1 := time.Now()
			h.ServeHTTP(lw, r)

			if lw.Status() == 0 {
				lw.WriteHeader(http.StatusOK)
			}

			entry.WithFields(logrus.Fields{
				"status": lw.Status(),
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
