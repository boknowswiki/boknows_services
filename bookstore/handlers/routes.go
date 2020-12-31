package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"www-github.cisco.com/bota/maglev-bootcamp/track-controlplane/week-1/bookstore/product"
)

// API add routes for the handlers
func API(client *mongo.Client, log *log.Logger) http.Handler {
	app := NewApp(log, product.Metrics())

	{
		c := Check{db: client}
		app.Handle(http.MethodGet, "/health", c.Health)
	}

	p := Products{DB: client, Log: log}

	app.Handle(http.MethodGet, "/books", p.List)
	app.Handle(http.MethodGet, "/books/{id}", p.Retrieve)
	app.Handle(http.MethodPost, "/books", p.Create)
	app.Handle(http.MethodPut, "/books/{id}", p.Update)
	app.Handle(http.MethodDelete, "/books/{id}", p.Delete)

	//app.Handle(http.MethodPost, "/v1/products/{id}/sales", p.AddSale)
	//app.Handle(http.MethodGet, "/v1/products/{id}/sales", p.ListSales)

	return app
}

// Handler is the signature used by all application handlers in this service.
type Handler func(http.ResponseWriter, *http.Request) error

// App is the entrypoint into our application and what controls the context of
// each request. Feel free to add any configuration data/logic on this type.
type App struct {
	log *log.Logger
	mux *chi.Mux
	mw  []product.Middleware
}

// NewApp constructs an App to handle a set of routes.
func NewApp(log *log.Logger, mw ...product.Middleware) *App {
	return &App{
		log: log,
		mux: chi.NewRouter(),
		mw:  mw,
	}
}

// Handle associates a handler function with an HTTP Method and URL pattern.
func (a *App) Handle(method, url string, h product.Handler) {
	h = product.WrapMiddleware(a.mw, h)

	fn := func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			a.log.Printf("ERROR: %+v", err)

			if err := RespondError(w, err); err != nil {
				a.log.Printf("ERROR: %v", err)
			}
		}
	}

	a.mux.MethodFunc(method, url, fn)
}

// Respond converts a Go value to JSON and sends it to the client.
func Respond(w http.ResponseWriter, data interface{}, statusCode int) error {

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	// Convert the response value to JSON.
	res, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Respond with the provided JSON.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if _, err := w.Write(res); err != nil {
		return err
	}

	return nil
}

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// ErrorResponse is the form used for API responses from failures in the API.
type ErrorResponse struct {
	Error  string       `json:"error"`
	Fields []FieldError `json:"fields,omitempty"`
}

// Error is used to pass an error during the request through the
// application with web specific context.
type Error struct {
	Err    error
	Status int
	Fields []FieldError
}

// NewRequestError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewRequestError(err error, status int) error {
	return &Error{err, status, nil}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (err *Error) Error() string {
	return err.Err.Error()
}

// RespondError sends an error reponse back to the client.
func RespondError(w http.ResponseWriter, err error) error {

	// If the error was of the type *Error, the handler has
	// a specific status code and error to return.
	if webErr, ok := errors.Cause(err).(*Error); ok {
		er := ErrorResponse{
			Error:  webErr.Err.Error(),
			Fields: webErr.Fields,
		}
		if err := Respond(w, er, webErr.Status); err != nil {
			return err
		}
		return nil
	}

	// If not, the handler sent any arbitrary error value so use 500.
	er := ErrorResponse{
		Error: http.StatusText(http.StatusInternalServerError),
	}
	if err := Respond(w, er, http.StatusInternalServerError); err != nil {
		return err
	}
	return nil
}

// ServeHTTP implements the http.Handler interface.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}
