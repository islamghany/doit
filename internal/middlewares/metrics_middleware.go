package middlewares

import (
	"net/http"

	"doit/internal/metrics"
	"doit/internal/web"
)

func Metrics() web.MiddleWare {
	return func(handler web.Handler) web.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			ctx := metrics.Set(r.Context())

			// val := web.GetValues(ctx1)

			err := handler(w, r.WithContext(ctx))

			n := metrics.AddRequest(ctx)

			// Add goroutines metric every 1000 requests to refresh the metric
			if n%1000 == 0 {
				metrics.AddGoroutines(ctx)
			}
			// status := web.GetStatusCode(ctx)
			// metrics.AddPromRequest(ctx, r.Method, r.URL.Path, fmt.Sprintf("%d", status))
			// metrics.AddPromLatency(ctx, r.Method, r.URL.Path, float64(time.Since(val.Time).Milliseconds()))
			metrics.AddGoroutines(ctx)
			if err != nil {
				metrics.AddError(ctx)
			}
			return err
		}
	}
}
