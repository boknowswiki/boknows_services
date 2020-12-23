package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/boknowswiki/boknows_services/garagesale/internal/platform/web"
	"github.com/boknowswiki/boknows_services/garagesale/internal/product"
)

// Products holds business logic related to Products.
type Products struct {
	DB  *sqlx.DB
	Log *log.Logger
}

// List gets all Products from the database then encodes them in a
// response to the client.
func (p *Products) List(w http.ResponseWriter, r *http.Request) error {
	list, err := product.List(p.DB)

	if err != nil {
		return errors.Wrap(err, "getting product list")
	}

	return web.Respond(w, list, http.StatusOK)
}

// Retrive gets a single Product from the database then encodes them in a
// response to the client.
func (p *Products) Retrive(w http.ResponseWriter, r *http.Request) error {

	id := chi.URLParam(r, "id")

	prod, err := product.Retrive(p.DB, id)

	if err != nil {
		switch err {
		case product.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case product.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "getting product %q", id)
		}
	}

	return web.Respond(w, prod, http.StatusOK)
}

// Create decodes the body of a request to create a new product. The full
// product with generated fields is sent back in the response.
func (p *Products) Create(w http.ResponseWriter, r *http.Request) error {
	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return errors.Wrap(err, "decoding product")
	}

	prod, err := product.Create(p.DB, np, time.Now())
	if err != nil {
		return errors.Wrap(err, "creating product")
	}

	return web.Respond(w, prod, http.StatusCreated)
}
