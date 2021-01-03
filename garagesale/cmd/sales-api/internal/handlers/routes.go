package handlers

import (
	"log"
	"net/http"

	"github.com/boknowswiki/boknows_services/garagesale/internal/mid"
	"github.com/boknowswiki/boknows_services/garagesale/internal/platform/auth"
	"github.com/boknowswiki/boknows_services/garagesale/internal/platform/web"
	"github.com/jmoiron/sqlx"
)

// API constructs an http.Handler with all application routes defined.
func API(db *sqlx.DB, log *log.Logger, authenticator *auth.Authenticator) http.Handler {

	app := web.NewApp(log, mid.Logger(log), mid.Errors(log), mid.Metrics())

	{
		c := Check{db: db}
		app.Handle(http.MethodGet, "/v1/health", c.Health)
	}

	u := Users{DB: db, authenticator: authenticator}

	app.Handle(http.MethodGet, "/v1/users/token", u.Token)

	p := Products{DB: db, Log: log}

	app.Handle(http.MethodGet, "/v1/products", p.List, mid.Authenticate(authenticator))
	app.Handle(http.MethodGet, "/v1/products/{id}", p.Retrieve, mid.Authenticate(authenticator))
	app.Handle(http.MethodPost, "/v1/products", p.Create, mid.Authenticate(authenticator))
	app.Handle(http.MethodPut, "/v1/products/{id}", p.Update, mid.Authenticate(authenticator))
	app.Handle(http.MethodDelete, "/v1/products/{id}", p.Delete, mid.Authenticate(authenticator))

	app.Handle(http.MethodPost, "/v1/products/{id}/sales", p.AddSale, mid.Authenticate(authenticator))
	app.Handle(http.MethodGet, "/v1/products/{id}/sales", p.ListSales, mid.Authenticate(authenticator))

	return app
}
