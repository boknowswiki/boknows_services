// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/boknowswiki/boknows_services/service/business/auth"
	"github.com/boknowswiki/boknows_services/service/business/data/user"
	"github.com/boknowswiki/boknows_services/service/business/mid"
	"github.com/boknowswiki/boknows_services/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, a *auth.Auth, db *sqlx.DB) *web.App {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	cg := checkGroup{
		build: build,
		//log:   log,
		db: db,
	}

	//app.Handle(http.MethodGet, "/readiness", check.readiness, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/readiness", cg.readiness)
	app.Handle(http.MethodGet, "/liveness", cg.liveness)

	// Register user management and authentication endpoints.
	ug := userGroup{
		user: user.New(log, db),
		auth: a,
	}
	app.Handle(http.MethodGet, "/users/:page/:rows", ug.query, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodPost, "/users", ug.create, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/users/:id", ug.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/users/:id", ug.update, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodDelete, "/users/:id", ug.delete, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/users/token/:kid", ug.token)

	/*
		// Register product and sale endpoints.
		pg := productGroup{
			product: product.New(log, db),
		}
		app.Handle(http.MethodGet, "/products", pg.query, mid.Authenticate(a))
		app.Handle(http.MethodPost, "/products", pg.create, mid.Authenticate(a))
		app.Handle(http.MethodGet, "/products/:id", pg.queryByID, mid.Authenticate(a))
		app.Handle(http.MethodPut, "/products/:id", pg.update, mid.Authenticate(a))
		app.Handle(http.MethodDelete, "/products/:id", pg.delete, mid.Authenticate(a))
	*/

	return app
}
