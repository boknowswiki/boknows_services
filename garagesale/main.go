package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	log.Printf("main: started")
	defer log.Println("main: finished")

	api := http.Server{
		Addr:         "localhost:8000",
		Handler:      http.HandlerFunc(Echo),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Starting to listen on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("error: listening and serving: %s", err)
	case sig := <-shutdown:
		log.Printf("main shuting down: %s", sig)

		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main gracefull shutdown in %v: %v", timeout, err)
			err = api.Close()
		}

		if err != nil {
			log.Fatalf("main: could not shutdown gracefully: %v", err)
		}

	}
}

// Echo is a basic HTTP Handler.
func Echo(w http.ResponseWriter, r *http.Request) {
	id := rand.Intn(1000)
	log.Printf("starting id: %v", id)
	defer log.Printf("end id: %v", id)

	time.Sleep(3 * time.Second)

	fmt.Fprintf(w, "You asked to %s %s\n", r.Method, r.URL.Path)
}
