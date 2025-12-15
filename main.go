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
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/observability-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	capiv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta1" //nolint:staticcheck // SA1019 deprecated package
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/giantswarm/logging-operator/internal/controller"
	"github.com/giantswarm/logging-operator/pkg/config"
	"github.com/giantswarm/logging-operator/pkg/resource"
	agentstoggle "github.com/giantswarm/logging-operator/pkg/resource/agents-toggle"
	credentials "github.com/giantswarm/logging-operator/pkg/resource/credentials"
	eventsloggerconfig "github.com/giantswarm/logging-operator/pkg/resource/events-logger-config"
	eventsloggersecret "github.com/giantswarm/logging-operator/pkg/resource/events-logger-secret"
	ingressauthsecret "github.com/giantswarm/logging-operator/pkg/resource/ingress-auth-secret"
	loggingconfig "github.com/giantswarm/logging-operator/pkg/resource/logging-config"
	loggingsecret "github.com/giantswarm/logging-operator/pkg/resource/logging-secret"
	loggingwiring "github.com/giantswarm/logging-operator/pkg/resource/logging-wiring"
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
	var enableNodeFiltering bool
	var enableTracing bool
	var enableNetworkMonitoring bool
	var includeEventsFromNamespaces StringSliceVar
	var excludeEventsFromNamespaces StringSliceVar
	var installationName string
	var customer string
	var pipeline string
	var region string
	var insecureCA bool
	var metricsAddr string
	var profilesAddr string
	var probeAddr string
	flag.Var(&defaultNamespaces, "default-namespaces", "List of namespaces to collect logs from by default on workload clusters")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableLogging, "enable-logging", true, "enable/disable logging for the whole installation")
	flag.BoolVar(&enableNodeFiltering, "enable-node-filtering", false, "enable/disable node filtering in Alloy logging configuration")
	flag.BoolVar(&enableTracing, "enable-tracing", false, "enable/disable tracing support for events logger")
	flag.BoolVar(&enableNetworkMonitoring, "enable-network-monitoring", false, "enable/disable network monitoring for the whole installation")
	flag.Var(&includeEventsFromNamespaces, "include-events-from-namespaces", "List of namespaces to collect events from on workload clusters (if empty, collect from all namespaces)")
	flag.Var(&excludeEventsFromNamespaces, "exclude-events-from-namespaces", "List of namespaces to exclude events from on workload clusters")
	flag.StringVar(&installationName, "installation-name", "unknown", "Name of the installation")
	flag.StringVar(&customer, "customer", "unknown", "Name of the customer")
	flag.StringVar(&pipeline, "pipeline", "unknown", "Name of the pipeline")
	flag.StringVar(&region, "region", "unknown", "Region where the installation is deployed")
	flag.BoolVar(&insecureCA, "insecure-ca", false, "Is the management cluter CA insecure?")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&profilesAddr, "pprof-bind-address", ":6060", "The address the pprof endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	opts := zap.Options{
		Development: false,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	discardHelmSecretsSelector, err := labels.Parse("owner notin (helm,Helm)")
	if err != nil {
		setupLog.Error(err, "failed to parse label selector")
		os.Exit(1)
	}

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
		Cache: cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&v1.Secret{}: {
					// Do not cache any helm secrets to reduce memory usage.
					Label: discardHelmSecretsSelector,
				},
			},
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create Config for dependency injection
	appConfig := config.Config{
		EnableLoggingFlag:           enableLogging,
		EnableNodeFilteringFlag:     enableNodeFiltering,
		EnableTracingFlag:           enableTracing,
		EnableNetworkMonitoringFlag: enableNetworkMonitoring,
		InstallationName:            installationName,
		Customer:                    customer,
		Pipeline:                    pipeline,
		Region:                      region,
		InsecureCA:                  insecureCA,
	}

	agentsToggle := agentstoggle.Resource{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	loggingWiring := loggingwiring.Resource{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	credentialSecrets := credentials.Resource{
		Client: mgr.GetClient(),
		Config: appConfig,
	}

	ingressAuthSecret := ingressauthsecret.Resource{
		Client: mgr.GetClient(),
		Config: appConfig,
	}

	loggingSecret := loggingsecret.Resource{
		Client: mgr.GetClient(),
		Config: appConfig,
	}

	loggingConfig := loggingconfig.Resource{
		Client:                           mgr.GetClient(),
		Config:                           appConfig,
		DefaultWorkloadClusterNamespaces: defaultNamespaces,
	}

	eventsLoggerConfig := eventsloggerconfig.Resource{
		Client:            mgr.GetClient(),
		Config:            appConfig,
		IncludeNamespaces: includeEventsFromNamespaces,
		ExcludeNamespaces: excludeEventsFromNamespaces,
	}

	eventsLoggerSecret := eventsloggersecret.Resource{
		Client: mgr.GetClient(),
		Config: appConfig,
	}

	if err = (&controller.CapiClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: appConfig,
		Resources: []resource.Interface{
			&agentsToggle,
			&loggingWiring,
			&credentialSecrets,
			&ingressAuthSecret,
			&loggingSecret,
			&loggingConfig,
			&eventsLoggerSecret,
			&eventsLoggerConfig,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create CAPI controller", "controller", "Cluster")
		os.Exit(1)
	}

	// The GrafanaOrganizationReconciler is only used in CAPI mode
	if err = (&controller.GrafanaOrganizationReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Resource: loggingConfig,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create GrafanaOrganization controller", "controller", "GrafanaOrganization")
		os.Exit(1)
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
