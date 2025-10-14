package web

import (
	"context"
	"time"
)

type valuesKey int

const key valuesKey = 1

type Values struct {
	TraceID    string
	Time       time.Time
	StatusCode int
}

func SetValues(ctx context.Context, values Values) context.Context {
	return context.WithValue(ctx, key, &values) // Store pointer
}

func GetValues(ctx context.Context) *Values {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return &Values{
			TraceID: "00000000-0000-0000-0000-000000000000",
			Time:    time.Now(),
		}
	}
	return v
}
func GetTraceID(ctx context.Context) string {
	return GetValues(ctx).TraceID
}

func GetTime(ctx context.Context) time.Time {
	return GetValues(ctx).Time
}

func GetStatusCode(ctx context.Context) int {
	return GetValues(ctx).StatusCode
}

// This means SetStatusCode won't work:
func SetStatusCode(ctx context.Context, statusCode int) {
	val, ok := ctx.Value(key).(*Values)
	if !ok {
		return
	}
	val.StatusCode = statusCode
}
