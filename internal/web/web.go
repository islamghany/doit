package web

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

type WebApp struct {
	mw []MiddleWare
	*http.ServeMux
}

func NewApp(mw ...MiddleWare) *WebApp {
	return &WebApp{
		mw:       mw,
		ServeMux: http.NewServeMux(),
	}
}

func (wa *WebApp) Handle(method, path string, handler Handler, mw ...MiddleWare) {
	handler = wrapMiddleware(handler, mw)
	handler = wrapMiddleware(handler, wa.mw)

	wa.handle(method, path, handler)
}

func (wa *WebApp) handle(method, path string, handler Handler) {
	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := SetValues(r.Context(), Values{
			TraceID: uuid.New().String(),
			Time:    time.Now(),
		})
		if err := handler(w, r.WithContext(ctx)); err != nil {
			log.Println(err)
		}
	}

	pattern := wa.parsePathForMux(method, path)
	wa.ServeMux.HandleFunc(pattern, h)
}
func (wa *WebApp) parsePathForMux(method, path string) string {
	// convert /api/v1/users/:id to  /api/v1/users/{id}
	// convert /api/v1/users/:id/:name to  /api/v1/users/{id}/{name}

	segments := strings.Split(path, "/")
	p := ""
	for i, segment := range segments {
		if segment != "" {
			if segment[0] == ':' {
				segment = "{" + segment[1:] + "}"
			}
			if i == 0 {
				p = segment
			} else {
				p = p + "/" + segment
			}
		}
	}

	return method + " " + p
}
