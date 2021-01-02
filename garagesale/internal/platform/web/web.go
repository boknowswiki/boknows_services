package web

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

// Handler is the signature used by all application handlers in this service.
type Handler func(http.ResponseWriter, *http.Request) error

// ctxKey represents the type of value for the context key.
type ctxKey int

// KeyValues is how request values or stored/retrieved.
const KeyValues ctxKey = 1

// Values carries information about each request.
type Values struct {
	StatusCode int
	Start      time.Time
}

// App is the entrypoint into our application and what controls the context of
// each request. Feel free to add any configuration data/logic on this type.
type App struct {
	log *log.Logger
	mux *chi.Mux
	mw  []Middleware
}

// NewApp constructs an App to handle a set of routes.
func NewApp(log *log.Logger, mw ...Middleware) *App {
	return &App{
		log: log,
		mux: chi.NewRouter(),
		mw:  mw,
	}
}

// Handle associates a handler function with an HTTP Method and URL pattern.
func (a *App) Handle(method, url string, h Handler) {

	// wrap the application's middleware around this endpoint's handler.
	h = wrapMiddleware(a.mw, h)

	fn := func(w http.ResponseWriter, r *http.Request) {
		// Create a Values struct to record state for the request. Store the
		// address in the request's context so it is sent down the call chain.
		v := Values{
			Start: time.Now(),
		}
		ctx := context.WithValue(r.Context(), KeyValues, &v)
		r = r.WithContext(ctx)

		err := h(w, r)
		if err != nil {
			a.log.Printf("Unhandled error %+v", err)
		}
	}

	a.mux.MethodFunc(method, url, fn)
}

// ServeHTTP implements the http.Handler interface.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}
