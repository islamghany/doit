package web

import (
	"context"
	"net/http"
)

const RequestIDHeader = "X-Request-ID"

type requestIDKey string

const requestIDKeyVal requestIDKey = "request_id"

func GetRequestIDHeader(r *http.Request) string {
	return r.Header.Get(RequestIDHeader)
}

func SetRequestIDForContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKeyVal, id)
}

func GetRequestIDFromContext(ctx context.Context) string {
	val, ok := ctx.Value(requestIDKeyVal).(string)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}
	return val
}
