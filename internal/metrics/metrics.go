package metrics

import (
	"expvar"
	"runtime"
	"sync/atomic"
)

var (
	requests   = expvar.NewInt("requests")   // current running goroutines
	errors     = expvar.NewInt("errors")     // total requests received.
	panics     = expvar.NewInt("panics")     // total errros occurred.
	goroutines = expvar.NewInt("goroutines") // total panics occurred.
	inFlight   = expvar.NewInt("in_flight")  // current in-flight requests.

	requestCount atomic.Int64 // For sampling check
)

func AddRequest() int64 {
	requests.Add(1)
	inFlight.Add(1)
	// Sample goroutines periodically (not every request)
	if requestCount.Add(1)%1000 == 0 {
		goroutines.Set(int64(runtime.NumGoroutine()))
	}
	return requests.Value()
}

func RequestDone() int64 {
	inFlight.Add(-1)
	return inFlight.Value()
}

func AddError() int64 {
	errors.Add(1)
	return errors.Value()
}

func AddPanics() int64 {
	panics.Add(1)
	return panics.Value()
}
