package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/boknowswiki/boknows_services/garagesale/internal/product"
)

// Products holds business logic related to Products.
type Products struct {
	DB *sqlx.DB
}

// List gets all Products from the database then encodes them in a
// response to the client.
func (p *Products) List(w http.ResponseWriter, r *http.Request) {
	list, err := product.List(p.DB)

	if err != nil {
		log.Printf("error: selecting products: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(list)
	if err != nil {
		log.Println("error marshalling result", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Println("error writing result", err)
	}
}
