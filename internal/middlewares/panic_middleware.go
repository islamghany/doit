package middlewares

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"doit/internal/metrics"
	"doit/internal/web"
)

func PanicMiddleware() web.MiddleWare {
	return func(handler web.Handler) web.Handler {
		return func(w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if recErr := recover(); recErr != nil {
					trace := debug.Stack()
					err = fmt.Errorf("panic [%v] trace[%s]", recErr, trace)
					metrics.AddPanics()
				}
			}()
			return handler(w, r)
		}
	}
}
