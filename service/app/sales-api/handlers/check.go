package handlers

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/boknowswiki/boknows_services/service/foundation/web"
)

type check struct {
	build string
	log   *log.Logger
}

func (c check) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	/*
		if n := rand.Intn(100); n%2 == 0 {
			//return errors.New("untrusted error")
			return web.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
			//panic("forcing panic")
			//return web.NewShutdownError("forcing shutdown")
		}
	*/
	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, w, status, http.StatusOK)

}

// liveness returns simple status info if the service is alive. If the
// app is deployed to a Kubernetes cluster, it will also return pod, node, and
// namespace details via the Downward API. The Kubernetes environment variables
// need to be set within your Pod/Deployment manifest.
func (c check) liveness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	/*
		ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.check.liveness")
		defer span.End()
	*/

	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	info := struct {
		Status    string `json:"status,omitempty"`
		Build     string `json:"build,omitempty"`
		Host      string `json:"host,omitempty"`
		Pod       string `json:"pod,omitempty"`
		PodIP     string `json:"podIP,omitempty"`
		Node      string `json:"node,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}{
		Status:    "up",
		Build:     c.build,
		Host:      host,
		Pod:       os.Getenv("KUBERNETES_PODNAME"),
		PodIP:     os.Getenv("KUBERNETES_NAMESPACE_POD_IP"),
		Node:      os.Getenv("KUBERNETES_NODENAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}

	return web.Respond(ctx, w, info, http.StatusOK)
}
