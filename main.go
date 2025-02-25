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
	"fmt"
	"os"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/observability-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/giantswarm/logging-operator/internal/controller"
	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	loggingreconciler "github.com/giantswarm/logging-operator/pkg/logging-reconciler"
	"github.com/giantswarm/logging-operator/pkg/reconciler"
	agentstoggle "github.com/giantswarm/logging-operator/pkg/resource/agents-toggle"
	eventsloggerconfig "github.com/giantswarm/logging-operator/pkg/resource/events-logger-config"
	eventsloggersecret "github.com/giantswarm/logging-operator/pkg/resource/events-logger-secret"
	grafanadatasource "github.com/giantswarm/logging-operator/pkg/resource/grafana-datasource"
	loggingconfig "github.com/giantswarm/logging-operator/pkg/resource/logging-config"
	loggingcredentials "github.com/giantswarm/logging-operator/pkg/resource/logging-credentials"
	loggingsecret "github.com/giantswarm/logging-operator/pkg/resource/logging-secret"
	loggingwiring "github.com/giantswarm/logging-operator/pkg/resource/logging-wiring"
	lokiingressauthsecret "github.com/giantswarm/logging-operator/pkg/resource/loki-ingress-auth-secret"
	proxyauth "github.com/giantswarm/logging-operator/pkg/resource/proxy-auth"
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
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

type StringSliceVar []string

func (s StringSliceVar) String() string {
	return strings.Join(s, ",")
}

func (s *StringSliceVar) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

func main() {
	var defaultNamespaces StringSliceVar
	var enableLeaderElection bool
	var enableLogging bool
	var loggingAgent string
	var eventsLogger string
	var installationName string
	var insecureCA bool
	var metricsAddr string
	var profilesAddr string
	var probeAddr string
	var vintageMode bool
	flag.Var(&defaultNamespaces, "default-namespaces", "List of namespaces to collect logs from by default on workload clusters")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableLogging, "enable-logging", true, "enable/disable logging for the whole installation")
	flag.StringVar(&loggingAgent, "logging-agent", common.LoggingAgentAlloy, fmt.Sprintf("select logging agent to use (%s or %s)", common.LoggingAgentPromtail, common.LoggingAgentAlloy))
	flag.StringVar(&eventsLogger, "events-logger", common.EventsLoggerAlloy, fmt.Sprintf("select events logger to use (%s or %s)", common.EventsLoggerAlloy, common.EventsLoggerGrafanaAgent))
	flag.StringVar(&installationName, "installation-name", "unknown", "Name of the installation")
	flag.BoolVar(&insecureCA, "insecure-ca", false, "Is the management cluter CA insecure?")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&profilesAddr, "pprof-bind-address", ":6060", "The address the pprof endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&vintageMode, "vintage", false, "Reconcile resources on a Vintage installation")
	opts := zap.Options{
		Development: false,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "5c8bbafe.x-k8s.io",
		PprofBindAddress:       profilesAddr,
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

	agentsToggle := agentstoggle.Reconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	loggingWiring := loggingwiring.Reconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	loggingSecrets := loggingcredentials.Reconciler{
		Client: mgr.GetClient(),
	}

	grafanaDatasource := grafanadatasource.Reconciler{
		Client: mgr.GetClient(),
	}

	proxyAuth := proxyauth.Reconciler{
		Client: mgr.GetClient(),
	}

	lokiIngressAuthSecret := lokiingressauthsecret.Reconciler{
		Client: mgr.GetClient(),
	}

	loggingSecret := loggingsecret.Reconciler{
		Client: mgr.GetClient(),
	}

	loggingConfig := loggingconfig.Reconciler{
		Client:                           mgr.GetClient(),
		DefaultWorkloadClusterNamespaces: defaultNamespaces,
	}

	eventsLoggerConfig := eventsloggerconfig.Reconciler{
		Client: mgr.GetClient(),
	}

	eventsLoggerSecret := eventsloggersecret.Reconciler{
		Client: mgr.GetClient(),
	}

	loggedcluster.O.EnableLoggingFlag = enableLogging
	loggedcluster.O.LoggingAgent = loggingAgent
	loggedcluster.O.KubeEventsLogger = eventsLogger
	loggedcluster.O.InstallationName = installationName
	loggedcluster.O.InsecureCA = insecureCA
	setupLog.Info("Loggedcluster config", "options", loggedcluster.O)

	loggingReconciler := loggingreconciler.LoggingReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Reconcilers: []reconciler.Interface{
			&agentsToggle,
			&loggingWiring,
			&loggingSecrets,
			&grafanaDatasource,
			&lokiIngressAuthSecret,
			&proxyAuth,
			&loggingSecret,
			&loggingConfig,
			&eventsLoggerSecret,
			&eventsLoggerConfig,
		},
	}

	if vintageMode {
		setupLog.Info("Vintage mode selected")

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
		setupLog.Info("CAPI mode selected")

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
