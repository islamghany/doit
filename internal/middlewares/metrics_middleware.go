package middlewares

import (
	"net/http"

	"doit/internal/metrics"
	"doit/internal/web"
)

func Metrics() web.MiddleWare {
	return func(handler web.Handler) web.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			metrics.AddRequest()
			defer metrics.RequestDone()

			// val := web.GetValues(ctx1)

			err := handler(w, r)
			// status := web.GetStatusCode(ctx)
			// metrics.AddPromRequest(ctx, r.Method, r.URL.Path, fmt.Sprintf("%d", status))
			// metrics.AddPromLatency(ctx, r.Method, r.URL.Path, float64(time.Since(val.Time).Milliseconds()))
			if err != nil {
				metrics.AddError()
			}
			return err
		}
	}
}
