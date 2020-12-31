package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	//"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "expvar"         // Register the expvar handlers
	_ "net/http/pprof" //Register the /debug/pprof handlers.

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/pkg/errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	//"go.mongodb.org/mongo-driver/bson/primitive"
	//"www-github.cisco.com/bota/maglev-bootcamp/track-controlplane/week-1/bookstore/conf"
	"www-github.cisco.com/bota/maglev-bootcamp/track-controlplane/week-1/bookstore/handlers"
)

// People ...
type People struct {
	First string
	Last  string
}

// Trainer ...
type Trainer struct {
	Name string
	Age  int
	City string
}

func main() {
	log.Println(os.Args)
	if err := run(); err != nil {
		log.Fatalf("error: run get : %v", err)
	}
}

func run() error {
	log.Println("start main")
	defer log.Println("end main")
	var url string
	var addr string
	var debugAddr string

	standalone := false

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "standalone":
			standalone = true
		default:
			log.Println("Unknown args: ", os.Args[1:])
		}
	}

	log := log.New(os.Stdout, "Bookstore : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	if standalone != true {
		kclient, err := k8s.NewInClusterClient()
		if err != nil {
			log.Fatal(err)
		}

		var nodes corev1.NodeList
		if err := kclient.List(context.Background(), "", &nodes); err != nil {
			log.Fatal(err)
		}
		for _, node := range nodes.Items {
			fmt.Printf("name=%q schedulable=%t\n", *node.Metadata.Name, !*node.Spec.Unschedulable)
		}

		var configMap corev1.ConfigMap
		err = kclient.Get(context.Background(), "default", "mongodb-configmap", &configMap)
		if err != nil {
			log.Fatalf("error failed to get configMap %v", err)
		}

		log.Printf("get configMap %#v", configMap)

		var mongodbSVC corev1.Service
		err = kclient.Get(context.Background(), "default", "mongodb-svc", &mongodbSVC)
		if err != nil {
			log.Fatalf("error failed to get svc %v", err)
		}

		//ip := *mongodbSVC.Spec.ClusterIP
		log.Printf("svc is %v", *mongodbSVC.Spec.ClusterIP)

		//a, _ := hex.DecodeString(ip)
		//fmt.Printf("%v.%v.%v.%v", a[3], a[2], a[1], a[0])

		url = fmt.Sprintf("mongodb://%s:27017", *mongodbSVC.Spec.ClusterIP)
		addr = "0.0.0.0:8888"
		debugAddr = "0.0.0.0:6666"
	} else {
		url = "mongodb://localhost:27017"
		addr = "127.0.0.1:8888"
		debugAddr = "127.0.0.1:6666"
	}
	log.Printf("getting new mongo client with url %v", url)
	mclient, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		log.Fatalf("error: failed to create client: %s", err)
	}

	log.Println("got new mongo client")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = mclient.Connect(ctx)

	defer func() {
		if err = mclient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	log.Println("client connet")

	// Start debug Service
	go func() {
		log.Println("debug service listening on", debugAddr)
		err := http.ListenAndServe(debugAddr, http.DefaultServeMux)
		log.Println("debug service closed", err)
	}()

	// Start API Service

	api := http.Server{
		Addr:         addr,
		Handler:      handlers.API(mclient, log),
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("main : API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "listening and serving")

	case <-shutdown:
		log.Println("main : Start shutdown")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main : Graceful shutdown did not complete in %v : %v", time.Second*5, err)
			err = api.Close()
		}

		if err != nil {
			return errors.Wrap(err, "main : could not stop server gracefully")
		}
	}

	return nil

}
