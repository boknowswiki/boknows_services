package product

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// BookCollector defines collections for prometheus.
type BookCollector struct {
	DB                   *mongo.Client
	BookCount            *prometheus.Desc
	BookGenreUniqueCount *prometheus.Desc
	BookInfo             *prometheus.Desc
}

//NewBookCollector returns a BookCollector instance.
func NewBookCollector(client *mongo.Client) *BookCollector {
	return &BookCollector{
		DB: client,
		BookCount: prometheus.NewDesc(
			"Bookstore_bookcount", "Shows bookstore number of books.", nil, nil,
		),
		BookGenreUniqueCount: prometheus.NewDesc(
			"Bookstore_genrecount", "Shows unique genre counts.", nil, nil,
		),
		BookInfo: prometheus.NewDesc(
			"Bookstore_bookinfo", "Shows books information.",
			[]string{"genre"}, nil,
		),
	}
}

// Describe sends the metrics to the channel.
func (bc *BookCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- bc.BookCount
	ch <- bc.BookGenreUniqueCount
	ch <- bc.BookInfo
}

// Collect gets the metrics and send to the channel.
func (bc *BookCollector) Collect(ch chan<- prometheus.Metric) {
	products := []Product{}

	collection := bc.DB.Database("test").Collection("books")

	ctx := context.Background()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Printf("find products error: %v", err)
		return
		//return nil, errors.Wrap(err, "selecting products")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var p Product
		if err = cursor.Decode(&p); err != nil {
			fmt.Printf("decode product error: %v", err)
			return
			//return nil, errors.Wrap(err, "decoding produces")
		}
		products = append(products, p)
	}

	productCount := len(products)
	genreMap := make(map[string]int)
	genreConut := 0

	for i := 0; i < productCount; i++ {
		prod := products[i]
		if _, ok := genreMap[prod.Genre]; !ok {
			genreMap[prod.Genre] = 1
			genreConut++
		} else {
			genreMap[prod.Genre]++
		}
	}

	for k, v := range genreMap {
		ch <- prometheus.MustNewConstMetric(bc.BookInfo, prometheus.GaugeValue, float64(v), k)
	}

	ch <- prometheus.MustNewConstMetric(bc.BookCount, prometheus.GaugeValue, float64(productCount))
	ch <- prometheus.MustNewConstMetric(bc.BookGenreUniqueCount, prometheus.GaugeValue, float64(genreConut))
	//ch <-
}
