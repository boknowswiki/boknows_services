package product_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"www-github.cisco.com/bota/maglev-bootcamp/track-controlplane/week-1/bookstore/product"
	"www-github.cisco.com/bota/maglev-bootcamp/track-controlplane/week-1/bookstore/tests"
)

// TestProducts tests product CRUD APIs.
func TestProducts(t *testing.T) {
	//Create mongodb container for testing.
	db, teardown := tests.NewUnit(t)
	defer teardown()

	ctx := context.Background()

	// Create NewProduct to test product.Create() function.
	newP1 := product.NewProduct{
		Name:   "Funny Book",
		Author: "Mike",
		ISBN:   "123456",
		Genre:  "funny",
	}
	now1 := time.Date(2021, time.January, 4, 0, 0, 0, 0, time.UTC)

	// Test product.Create() function.
	p0, err := product.Create(ctx, db, newP1, now1)
	if err != nil {
		t.Fatalf("creating product p0: %s", err)
	}

	// Test product.Retrieve() function.
	p1, err := product.Retrieve(ctx, db, p0.ID)
	if err != nil {
		t.Fatalf("getting product p0: %s", err)
	}

	// Compare the retrieved product is the same as created.
	if diff := cmp.Diff(p1, p0); diff != "" {
		t.Fatalf("fetched != created:\n%s", diff)
	}

	// Create NewProduct to test product.Create() and product.List() function.
	newP2 := product.NewProduct{
		Name:   "Fiction Book",
		Author: "Mike",
		ISBN:   "345678",
		Genre:  "fiction",
	}

	now2 := time.Date(2021, time.January, 5, 0, 0, 0, 0, time.UTC)
	p2, err := product.Create(ctx, db, newP2, now2)
	if err != nil {
		t.Fatalf("creating product p0: %s", err)
	}

	p3, err := product.Retrieve(ctx, db, p2.ID)
	if err != nil {
		t.Fatalf("getting product p0: %s", err)
	}

	if diff := cmp.Diff(p2, p3); diff != "" {
		t.Fatalf("fetched != created:\n%s", diff)
	}

	// Test case for product.Update() function.
	updateAuthor := "Ben"
	updateP4 := product.UpdateProduct{
		Author: &updateAuthor,
	}
	if err := product.Update(ctx, db, p2.ID, updateP4, now2); err != nil {
		t.Fatalf("updating product %v: %s", updateP4, err)
	}

	p3, err = product.Retrieve(ctx, db, p2.ID)
	if err != nil {
		t.Fatalf("getting product p0: %s", err)
	}

	if exp, got := "Ben", p3.Author; exp != got {
		t.Fatalf("expected product %v, got %v", exp, got)
	}

	// Test case for product.List() function.
	ps, err := product.List(ctx, db)
	if err != nil {
		t.Fatalf("listing products: %s", err)
	}
	if exp, got := 2, len(ps); exp != got {
		t.Fatalf("expected product list size %v, got %v", exp, got)
	}

	if err = product.Delete(ctx, db, p0.ID); err != nil {
		log.Fatalf("deleting product %v: %s", p0.ID, err)
	}

	if err = product.Delete(ctx, db, p2.ID); err != nil {
		log.Fatalf("deleting product %v: %s", p2.ID, err)
	}

	ps, err = product.List(ctx, db)
	if err != nil {
		t.Fatalf("listing products: %s", err)
	}
	if exp, got := 0, len(ps); exp != got {
		t.Fatalf("expected product list size %v, got %v", exp, got)
	}
}
