package handlers

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// Check is for health check.
type Check struct {
	db *mongo.Client
}

// Health validates the service is healthy and ready to accept requests.
func (c *Check) Health(w http.ResponseWriter, r *http.Request) error {

	var health struct {
		Status string `json:"status"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.db.Ping(ctx, nil)
	if err != nil {
		health.Status = "db not ready"
		return Respond(w, health, http.StatusInternalServerError)
	}

	health.Status = "ok"
	return Respond(w, health, http.StatusOK)
}
