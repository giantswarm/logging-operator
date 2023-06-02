package promtailtoggle

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
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
	logger.Info("promtailtoggle create")

	desiredConfigMap, err := GenerateObservabilityBundleConfigMap(cluster)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info(fmt.Sprintf("promtailtoggle checking %s/%s", desiredConfigMap.GetNamespace(), desiredConfigMap.GetName()))
	var currentConfigMap v1.ConfigMap
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredConfigMap.GetName(), Namespace: desiredConfigMap.GetNamespace()}, &currentConfigMap)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("promtailtoggle not found, creating")
			err = r.Client.Create(ctx, &desiredConfigMap)
		}
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	}

	if needUpdate(currentConfigMap, desiredConfigMap) {
		logger.Info("promtailtoggle updating")
		err := r.Client.Update(ctx, &desiredConfigMap)
		if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info("promtailtoggle up to date")
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) ReconcileDelete(ctx context.Context, cluster capiv1beta1.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("promtailtoggle delete")

	desiredConfigMap, err := GenerateObservabilityBundleConfigMap(cluster)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info(fmt.Sprintf("promtailtoggle deleting %s/%s", desiredConfigMap.GetNamespace(), desiredConfigMap.GetName()))
	err = r.Client.Delete(ctx, &desiredConfigMap)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info("promtailtoggle already deleted")
		} else if err != nil {
			return ctrl.Result{}, errors.WithStack(err)
		}
	} else {
		logger.Info("promtailtoggle deleted")
	}

	return ctrl.Result{}, nil
}

func needUpdate(current, desired v1.ConfigMap) bool {
	return !reflect.DeepEqual(current.Data, desired.Data)
}
