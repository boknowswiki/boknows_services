package product

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	//"www-github.cisco.com/bota/maglev-bootcamp/track-controlplane/week-1/bookstore/product"
)

// Product is an item we sell.
type Product struct {
	ID          string    `db:"product_id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Author      string    `db:"author" json:"author"`
	ISBN        string    `db:"isbn" json:"isbn"`
	Genre       string    `db:"genre" json:"genre`
	DateCreated time.Time `db:"datecreated" json:"date_created"`
	DateUpdated time.Time `db:"dateupdated" json:"date_updated"`
}

// NewProduct get new product from user.
type NewProduct struct {
	Name   string `db:"name" json:"name" validate:"required"`
	Author string `db:"author" json:"author"`
	ISBN   string `db:"isbn" json:"isbn"`
	Genre  string `db:"genre" json:"genre"`
}

// UpdateProduct defines what information may be provided to modify an
// existing Product. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateProduct struct {
	Name   *string `json:"name"`
	Author *string `json:"author"`
	ISBN   *string `json:"isbn"`
	Genre  *string `json:"genre`
}

// Predefined errors identify expected failure conditions.
var (
	// ErrNotFound is used when a specific Product is requested but does not exist.
	ErrNotFound = errors.New("product not found")

	// ErrInvalidID is used when an invalid UUID is provided.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// List gets all Products from the database.
func List(ctx context.Context, db *mongo.Client) ([]Product, error) {
	products := []Product{}

	collection := db.Database("test").Collection("books")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "selecting products")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var p Product
		if err = cursor.Decode(&p); err != nil {
			return nil, errors.Wrap(err, "decoding produces")
		}
		products = append(products, p)
	}

	return products, nil
}

// Retrieve gets a single Product from the database.
func Retrieve(ctx context.Context, db *mongo.Client, id string) (*Product, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var p Product

	collection := db.Database("test").Collection("books")

	filter := bson.D{{"id", id}}

	err := collection.FindOne(ctx, filter).Decode(&p)
	if err != nil {
		return nil, errors.Wrap(err, "get product")
	}

	return &p, nil
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated..
func Create(ctx context.Context, db *mongo.Client, np NewProduct, now time.Time) (*Product, error) {
	p := Product{
		ID:          uuid.New().String(),
		Name:        np.Name,
		Author:      np.Author,
		ISBN:        np.ISBN,
		Genre:       np.Genre,
		DateCreated: now.UTC(),
		DateUpdated: now.UTC(),
	}

	collection := db.Database("test").Collection("books")

	_, err := collection.InsertOne(ctx, p)
	if err != nil {
		return nil, errors.Wrap(err, "inserting product")
	}

	return &p, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func Update(ctx context.Context, db *mongo.Client, id string, update UpdateProduct, now time.Time) error {
	p, err := Retrieve(ctx, db, id)
	if err != nil {
		return err
	}

	if update.Name != nil {
		p.Name = *update.Name
	}

	if update.Author != nil {
		p.Author = *update.Author
	}
	if update.ISBN != nil {
		p.ISBN = *update.ISBN
	}
	if update.Genre != nil {
		p.Genre = *update.Genre
	}

	p.DateUpdated = now

	collection := db.Database("test").Collection("books")

	filter := bson.D{{"id", id}}

	_, err = collection.UpdateOne(ctx, filter,
		bson.D{
			{"$set", bson.D{
				{"name", p.Name},
				{"author", p.Author},
				{"isbn", p.ISBN},
				{"grene", p.DateUpdated},
				{"dateupdated", p.DateUpdated},
			}},
		})
	if err != nil {
		return errors.Wrap(err, "get product")
	}

	return nil
}

// Delete removes the product identified by a given ID.
func Delete(ctx context.Context, db *mongo.Client, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	collection := db.Database("test").Collection("books")

	filter := bson.D{{"id", id}}

	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "delete product")
	}

	return nil
}
