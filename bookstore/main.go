package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "expvar"         // Register the expvar handlers
	_ "net/http/pprof" //Register the /debug/pprof handlers.

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"www-github.cisco.com/bota/maglev-bootcamp/track-controlplane/week-1/bookstore/handlers"
	"www-github.cisco.com/bota/maglev-bootcamp/track-controlplane/week-1/bookstore/product"
)

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

	// Standalone is for running without minikube, with docker-compose.
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

	// For minikube case to get mongodb url.
	if standalone != true {
		// Get mongodb service url from environments.
		mongoSVC := os.Getenv("MONGODB_SVC_PORT_27017_TCP")
		if mongoSVC == "" {
			err := errors.New("failed to get mongo service")
			return err

		}
		mongoSVC = strings.Replace(mongoSVC, "tcp", "mongodb", -1)
		log.Println("mongo url: ", mongoSVC)

		// Below gets the mongodb service url from configmap.
		/*
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

			//log.Printf("svc is %v", *mongodbSVC.Spec.ClusterIP)

			url = fmt.Sprintf("mongodb://%s:27017", *mongodbSVC.Spec.ClusterIP)
		*/
		url = mongoSVC
		addr = "0.0.0.0:8888"
		debugAddr = "0.0.0.0:6060"
	} else {
		url = "mongodb://localhost:27017"
		addr = "127.0.0.1:8888"
		debugAddr = "127.0.0.1:6060"
	}

	log.Printf("getting new mongo client with url %v", url)
	mclient, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		err = errors.Wrap(err, "failed to create client")
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = mclient.Connect(ctx)

	defer func() {
		if err = mclient.Disconnect(ctx); err != nil {
			log.Printf("mongodb disconnect failed: %v", err)
		}
		cancel()
	}()

	// Start metrics
	go func() {
		bc := product.NewBookCollector(mclient)
		prometheus.MustRegister(bc)

		log.Println("prometheus metric on 2112")
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

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

	// Make a channel to get server errors.
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
