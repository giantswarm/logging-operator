/*
Copyright 2023.

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
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"

	"github.com/giantswarm/logging-operator/internal/controller"
	loggingreconciler "github.com/giantswarm/logging-operator/pkg/logging-reconciler"
	"github.com/giantswarm/logging-operator/pkg/reconciler"
	grafanadatasource "github.com/giantswarm/logging-operator/pkg/resource/grafana-datasource"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
	lokiauth "github.com/giantswarm/logging-operator/pkg/resource/loki-auth"
	promtailtoggle "github.com/giantswarm/logging-operator/pkg/resource/promtail-toggle"
	promtailwiring "github.com/giantswarm/logging-operator/pkg/resource/promtail-wiring"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(capiv1beta1.AddToScheme(scheme))
	utilruntime.Must(appv1.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func main() {
	var enableLeaderElection bool
	var metricsAddr string
	var probeAddr string
	var vintageMode bool
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&vintageMode, "vintage", false, "Reconcile resources on a Vintage installation")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "5c8bbafe.x-k8s.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	promtailReconciler := promtailtoggle.Reconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	promtailWiring := promtailwiring.Reconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	loggingSecrets := loggingcredentials.Reconciler{
		Client: mgr.GetClient(),
	}

	grafanaDatasource := grafanadatasource.Reconciler{
		Client: mgr.GetClient(),
	}

	lokiAuth := lokiauth.Reconciler{
		Client: mgr.GetClient(),
	}

	loggingReconciler := loggingreconciler.LoggingReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Reconcilers: []reconciler.Interface{
			&promtailReconciler,
			&promtailWiring,
			&loggingSecrets,
			&grafanaDatasource,
			&lokiAuth,
		},
	}

	if vintageMode {

		if err = (&controller.VintageMCReconciler{
			Client:            mgr.GetClient(),
			Scheme:            mgr.GetScheme(),
			LoggingReconciler: loggingReconciler,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create Vintage MC controller", "controller", "Service")
			os.Exit(1)
		}
		if err = (&controller.VintageWCReconciler{
			Client:            mgr.GetClient(),
			Scheme:            mgr.GetScheme(),
			LoggingReconciler: loggingReconciler,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create Vintage WC controller", "controller", "Service")
			os.Exit(1)
		}

	} else {

		if err = (&controller.CapiClusterReconciler{
			Client:            mgr.GetClient(),
			Scheme:            mgr.GetScheme(),
			LoggingReconciler: loggingReconciler,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create CAPI controller", "controller", "Cluster")
			os.Exit(1)
		}
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
