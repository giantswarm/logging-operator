/*
Copyright 2024.

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

package controller

import (
	"context"
	"time"

	grafanaorganization "github.com/giantswarm/observability-operator/api/v1alpha1"
	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
	"github.com/giantswarm/logging-operator/pkg/logged-cluster/capicluster"
	loggingconfig "github.com/giantswarm/logging-operator/pkg/resource/logging-config"
)

// GrafanaOrganizationReconciler reconciles grafanaOrganization CRs
type GrafanaOrganizationReconciler struct {
	Client                  client.Client
	Scheme                  *runtime.Scheme
	Reconciler              loggingconfig.Reconciler
	ManagementClusterConfig common.ManagementClusterConfig
}

//+kubebuilder:rbac:groups=observability.giantswarm.io,resources=grafanaorganizations,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=observability.giantswarm.io,resources=grafanaorganizations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// It is triggered whenever a change happen to a grafanaOrganization CR and
// calls the logging config reconciler for each cluster so that their tenant
// list is always up to date.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (g *GrafanaOrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	logger := log.FromContext(ctx)

	logger.Info("Started reconciling Grafana Organization")
	defer logger.Info("Finished reconciling Grafana Organization")

	clusters := &capi.ClusterList{}
	err = g.Client.List(ctx, clusters)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.WithStack(err)
	}

	for _, cluster := range clusters.Items {
		loggedCluster := &capicluster.Object{
			Object: &cluster,
			LoggingAgent: &loggedcluster.LoggingAgent{
				LoggingAgent:     g.ManagementClusterConfig.DefaultLoggingAgent,
				KubeEventsLogger: g.ManagementClusterConfig.DefaultKubeEventsLogger,
			},
		}

		if common.IsLoggingEnabled(g.ManagementClusterConfig, loggedCluster) {
			err = common.ToggleAgents(ctx, g.Client, loggedCluster)
			if err != nil {
				// Handle case where the app is not found.
				if apimachineryerrors.IsNotFound(err) {
					logger.Info("observability bundle app not found, requeueing")
					// If the app is not found we should requeue and try again later (5 minutes is the app platform default reconciliation time)
					return ctrl.Result{RequeueAfter: time.Duration(5 * time.Minute)}, nil
				}
				return ctrl.Result{}, errors.WithStack(err)
			}

			// Reconcile logging config for each cluster
			result, err := g.Reconciler.ReconcileCreate(ctx, loggedCluster)
			if err != nil {
				return result, errors.WithStack(err)
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (g *GrafanaOrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanaorganization.GrafanaOrganization{}).
		Complete(g)
}
