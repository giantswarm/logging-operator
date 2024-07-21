package common

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Desired interface {
	v1.ConfigMap | v1.Secret
}

type DesiredPtr[T any] interface {
	client.Object
	*T
}

func PointerTo[T any](v T) *T {
	return &v
}

func Ensure[T Desired, PT DesiredPtr[T]](ctx context.Context, client client.Client, desiredResource T, needUpdate func(T, T) bool, reconcilerName string) error {
	var (
		currentResource    T
		currentResourcePtr = PT(&currentResource)
		desiredResourcePtr = PT(&desiredResource)
	)
	logger := log.FromContext(ctx)

	// Check if config already exists.
	logger.Info(fmt.Sprintf("%s - getting", reconcilerName), "namespace", desiredResourcePtr.GetNamespace(), "name", desiredResourcePtr.GetName())
	err := client.Get(ctx, types.NamespacedName{Name: desiredResourcePtr.GetName(), Namespace: desiredResourcePtr.GetNamespace()}, currentResourcePtr)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("%s - creating", reconcilerName), "namespace", desiredResourcePtr.GetNamespace(), "name", desiredResourcePtr.GetName())
			err = client.Create(ctx, desiredResourcePtr)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		}
		return errors.WithStack(err)
	}

	if !needUpdate(currentResource, desiredResource) {
		logger.Info(fmt.Sprintf("%s - up to date", reconcilerName), "namespace", desiredResourcePtr.GetNamespace(), "name", desiredResourcePtr.GetName())
		return nil
	}

	logger.Info(fmt.Sprintf("%s -updating", reconcilerName), "namespace", desiredResourcePtr.GetNamespace(), "name", desiredResourcePtr.GetName())
	err = client.Update(ctx, desiredResourcePtr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
