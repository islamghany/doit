package web

type MiddleWare func(Handler) Handler

func wrapMiddleware(handler Handler, mw []MiddleWare) Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		mwFnc := mw[i]
		if mwFnc != nil {
			handler = mwFnc(handler)
		}
	}
	return handler
}
