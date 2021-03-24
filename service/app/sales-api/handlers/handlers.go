// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/boknowswiki/boknows_services/service/business/auth"
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
	return app
}
