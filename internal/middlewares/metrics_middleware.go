package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"doit/internal/metrics"
	"doit/internal/web"
)

func Metrics() web.MiddleWare {
	return func(handler web.Handler) web.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			start := time.Now()

			// Track in-flight for Prometheus
			metrics.HTTPRequestsInFlight.Inc()
			defer metrics.HTTPRequestsInFlight.Dec()

			metrics.AddRequest()
			defer metrics.RequestDone()

			// val := web.GetValues(ctx1)

			err := handler(w, r)
			duration := time.Since(start).Seconds()
			status := web.GetStatusCode(r.Context())

			metrics.RecordHTTPRequest(r.Method, r.URL.Path, fmt.Sprintf("%d", status), duration)

			if err != nil {
				metrics.AddError()
			}
			return err
		}
	}
}
