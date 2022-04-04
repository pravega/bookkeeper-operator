/*
Copyright 2022.

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
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/operator-framework/operator-lib/leader"

	bookkeeperv1alpha1 "github.com/pravega/bookkeeper-operator/api/v1alpha1"
	"github.com/pravega/bookkeeper-operator/controllers"

	//controllerconfig "github.com/pravega/bookkeeper-operator/pkg/controller/config"
	controllerconfig "github.com/pravega/bookkeeper-operator/pkg/controller/config"
	"github.com/pravega/bookkeeper-operator/pkg/version"
	"github.com/sirupsen/logrus"
	//+kubebuilder:scaffold:imports
)

var (
	versionFlag bool
	webhookFlag bool
	log         = ctrl.Log.WithName("cmd")
	scheme      = apimachineryruntime.NewScheme()
)

func init() {
	flag.BoolVar(&versionFlag, "version", false, "Show version and quit")
	flag.BoolVar(&controllerconfig.TestMode, "test", false, "Enable test mode. Do not use this flag in production")
	flag.BoolVar(&webhookFlag, "webhook", true, "Enable webhook, the default is enabled.")
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(bookkeeperv1alpha1.AddToScheme(scheme))
}

func printVersion() {
	log.Info(fmt.Sprintf("zookeeper-operator Version: %v", version.Version))
	log.Info(fmt.Sprintf("Git SHA: %s", version.GitSHA))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", "127.0.0.1:6000", "The address the metric endpoint binds to.")

	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))

	printVersion()

	if versionFlag {
		os.Exit(0)
	}

	if controllerconfig.TestMode {
		logrus.Warn("----- Running in test mode. Make sure you are NOT in production -----")
	}

	if controllerconfig.DisableFinalizer {
		logrus.Warn("----- Running with finalizer disabled. -----")
	}
	/*
			namespace, err := k8sutil.GetWatchNamespace()
			if err != nil {
				log.Fatal(err, "failed to get watch namespace")
			}

		// Get a config to talk to the apiserver
		cfg, err := config.GetConfig()
		if err != nil {
			logrus.Fatal(err)
		}

		operatorNs, err := k8sutil.GetOperatorNamespace()
			if err != nil {
				log.Error(err, "failed to get operator namespace")
				os.Exit(1)
			}*/

	ctx := context.TODO()
	// Become the leader before proceeding
	err := leader.Become(ctx, "bookkeeper-operator-lock")
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	log.Info("Registering Components")

	if err = (&controllers.BookkeeperClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		//	Log:    ctrl.Log.WithName("controllers").WithName("BookkeeperCluster"),
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "BookkeeperCluster")
		os.Exit(1)
	}
	if webhookFlag {
		if err = (&bookkeeperv1alpha1.BookkeeperCluster{}).SetupWebhookWithManager(mgr); err != nil {
			log.Error(err, "unable to create webhook", "webhook", "BookkeeperCluster")
			os.Exit(1)
		}
	}
	//+kubebuilder:scaffold:builder

	log.Info("Webhook Setup completed")
	log.Info("Starting the Cmd")

	// Start the Cmd

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
