/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	booksv1 "www-github.cisco.com/bota/maglev-bootcamp/track-controlplane/week-3/bookstore-operator/api/v1"
	"www-github.cisco.com/bota/maglev-bootcamp/track-controlplane/week-3/bookstore-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(booksv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "16160df9.bookstore.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	/*
		env := os.Environ()
		for _, e := range env {
			setupLog.Info(e)
		}
	*/

	bookstoreServiceHost := os.Getenv("BOOKSTORE_SERVICE_SERVICE_HOST")
	if bookstoreServiceHost == "" {
		setupLog.Error(err, "unable to get bookstore service host", "controller", "Book")
		os.Exit(1)
	}

	bookstoreServicePort := os.Getenv("BOOKSTORE_SERVICE_SERVICE_PORT")
	if bookstoreServicePort == "" {
		setupLog.Error(err, "unable to get bookstore service port", "controller", "Book")
		os.Exit(1)
	}

	bookstoreService := fmt.Sprintf("http://%s:%s", bookstoreServiceHost, bookstoreServicePort)

	log.Println("get bookstore service: ", bookstoreService)

	if err = (&controllers.BookReconciler{
		Client: mgr.GetClient(),
		//Log:    ctrl.Log.WithName("controllers").WithName("Book"),
		Log:    log.New(os.Stdout, "Bookstore : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile),
		Scheme: mgr.GetScheme(),
		SVC:    bookstoreService,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Book")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
