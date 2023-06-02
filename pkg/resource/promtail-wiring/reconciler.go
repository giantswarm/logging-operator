package promtailwiring

import (
	"context"
	"fmt"

	appv1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	promtailtoggle "github.com/giantswarm/logging-operator/pkg/resource/promtail-toggle"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *Reconciler) ReconcileCreate(ctx context.Context, cluster capiv1beta1.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailwiring create")

	appMeta := ObservabilityBundleAppMeta(cluster)

	logger.Info(fmt.Sprintf("promtailwiring checking %s/%s", appMeta.GetNamespace(), appMeta.GetNamespace()))
	var currentApp appv1.App
	err := r.Client.Get(ctx, types.NamespacedName{Name: appMeta.GetName(), Namespace: appMeta.GetNamespace()}, &currentApp)
	if err != nil {
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if setUserConfig(&currentApp, cluster) {
		logger.Info("promtailwiring updating")
		err := r.Client.Update(ctx, &currentApp)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info("promtailwiring up to date")
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) ReconcileDelete(ctx context.Context, cluster capiv1beta1.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailwiring delete")

	return ctrl.Result{}, nil
}

func setUserConfig(app *appv1.App, cluster capiv1beta1.Cluster) bool {
	observabilityBundleConfigMapMeta := promtailtoggle.ObservabilityBundleConfigMapMeta(cluster)
	updated := app.Spec.UserConfig.ConfigMap.Name != observabilityBundleConfigMapMeta.GetName() || app.Spec.UserConfig.ConfigMap.Namespace == observabilityBundleConfigMapMeta.GetNamespace()

	app.Spec.UserConfig.ConfigMap.Name = observabilityBundleConfigMapMeta.GetName()
	app.Spec.UserConfig.ConfigMap.Namespace = observabilityBundleConfigMapMeta.GetNamespace()

	return updated
}
