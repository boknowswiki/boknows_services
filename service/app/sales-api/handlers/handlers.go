// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/boknowswiki/boknows_services/service/foundation/web"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger) *web.App {
	app := web.NewApp(shutdown)

	check := check{
		log: log,
	}

	app.Handle(http.MethodGet, "/readiness", check.readiness)
	return app
}
