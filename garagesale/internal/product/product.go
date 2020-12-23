package product

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// List gets all Products from the database.
func List(db *sqlx.DB) ([]Product, error) {
	products := []Product{}

	const q = `SELECT * FROM products`

	if err := db.Select(&products, q); err != nil {
		return nil, errors.Wrap(err, "selecting products")
	}

	return products, nil
}

// Retrive gets a single Product from the database.
func Retrive(db *sqlx.DB, id string) (*Product, error) {
	var p Product

	const q = `SELECT * FROM products WHERE product_id = $1`

	if err := db.Get(&p, q, id); err != nil {
		return nil, errors.Wrap(err, "get product")
	}

	return &p, nil
}
