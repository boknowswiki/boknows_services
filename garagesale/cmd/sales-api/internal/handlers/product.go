package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"

	"github.com/boknowswiki/boknows_services/garagesale/internal/product"
)

// Products holds business logic related to Products.
type Products struct {
	DB  *sqlx.DB
	Log *log.Logger
}

// List gets all Products from the database then encodes them in a
// response to the client.
func (p *Products) List(w http.ResponseWriter, r *http.Request) {
	list, err := product.List(p.DB)

	if err != nil {
		p.Log.Printf("error: selecting products: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(list)
	if err != nil {
		p.Log.Println("error marshalling result", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		p.Log.Println("error writing result", err)
	}
}

// Retrive gets a single Product from the database then encodes them in a
// response to the client.
func (p *Products) Retrive(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	product, err := product.Retrive(p.DB, id)

	if err != nil {
		p.Log.Printf("error: selecting products: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(product)
	if err != nil {
		p.Log.Println("error marshalling result", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		p.Log.Println("error writing result", err)
	}
}

// Create decodes the body of a request to create a new product. The full
// product with generated fields is sent back in the response.
func (p *Products) Create(w http.ResponseWriter, r *http.Request) {
	var np product.NewProduct
	if err := json.NewDecoder(r.Body).Decode(&np); err != nil {
		p.Log.Println("decoding product", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	prod, err := product.Create(p.DB, np, time.Now())
	if err != nil {
		p.Log.Println("creating product", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(prod)
	if err != nil {
		p.Log.Println("error marshalling result", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		p.Log.Println("error writing result", err)
	}
}
