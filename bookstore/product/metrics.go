package product

import (
	"expvar"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	//"github.com/struCoder/pidusage"
)

// m contains the global program counters for the application.
var m = struct {
	gr  *expvar.Int
	req *expvar.Int
	err *expvar.Int
}{
	gr:  expvar.NewInt("goroutines"),
	req: expvar.NewInt("requests"),
	err: expvar.NewInt("errors"),
}

var (
	bookstoreResponseLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "Bookstore_response_latency",
			Help: "The http response latency",
			//ConstLabels: "Method",
			Buckets: []float64{1, 2, 5, 10, 20, 60},
		},
		[]string{"GET"},
	)

	bookstoreRequestNumber = promauto.NewCounter(prometheus.CounterOpts{
		Name: "Bookstore_request_number",
		Help: "The http request number",
	})
	bookstoreErrorNumber = promauto.NewCounter(prometheus.CounterOpts{
		Name: "Bookstore_error_number",
		Help: "The http error number",
	})
	bookstoreSuccessResponseNumber = promauto.NewCounter(prometheus.CounterOpts{
		Name: "Bookstore_success_response_number",
		Help: "The success response number",
	})
)

// Metrics updates program counters.
func Metrics() Middleware {
	// This is the actual middleware function to be executed.
	f := func(before Handler) Handler {

		// Wrap this handler around the next one provided.
		h := func(w http.ResponseWriter, r *http.Request) error {

			start := time.Now()

			err := before(w, r)

			duration := time.Since(start)

			bookstoreResponseLatency.WithLabelValues(r.Method).Observe(float64(duration.Milliseconds()))
			// Increment the request counter.
			m.req.Add(1)
			//fmt.Println("url path: ", r.URL.Path)
			if r.URL.Path == "/books" {
				bookstoreRequestNumber.Inc()
			}

			// Update the count for the number of active goroutines every 100 requests.
			if m.req.Value()%100 == 0 {
				m.gr.Set(int64(runtime.NumGoroutine()))
			}

			// Increment the errors counter if an error occurred on this request.
			if err != nil {
				m.err.Add(1)
				bookstoreErrorNumber.Inc()
			} else { // Increment success response number.
				if r.URL.Path == "/books" {
					bookstoreSuccessResponseNumber.Inc()
				}
			}

			// Return the error so it can be handled further up the chain.
			return err
		}

		return h
	}

	return f
}
