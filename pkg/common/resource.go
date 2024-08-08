package common

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	apimachineryerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Generic constaint to handle pointer type.
// This constaint allow to use a pointer to T as generic type.
// This contraint inherits from client.Object as we expect
// pointer to T to satisfy the client.Object interface.
type Ptr[T any] interface {
	client.Object
	*T
}

// EnsureCreatedOrUpdated ensure the desiredResource exists and is up to date in the Kubernetes API.
// If the resource does not exist, it will be created.
// If the resource exists but is not up to date, it will be updated when needUpdate returns true.
func EnsureCreatedOrUpdated[T any, PT Ptr[T]](ctx context.Context, client client.Client, desiredResource T, needUpdate func(T, T) bool, reconcilerName string) error {
	var (
		currentResource    T
		currentResourcePtr = PT(&currentResource)
		desiredResourcePtr = PT(&desiredResource)
	)

	logger := log.FromContext(ctx)

	// Check if resource already exists.
	logger.Info(fmt.Sprintf("%s - getting", reconcilerName), "namespace", desiredResourcePtr.GetNamespace(), "name", desiredResourcePtr.GetName())
	err := client.Get(ctx, types.NamespacedName{Name: desiredResourcePtr.GetName(), Namespace: desiredResourcePtr.GetNamespace()}, currentResourcePtr)
	if err != nil {
		if apimachineryerrors.IsNotFound(err) {
			// Resource was not found, create it.
			logger.Info(fmt.Sprintf("%s - creating", reconcilerName), "namespace", desiredResourcePtr.GetNamespace(), "name", desiredResourcePtr.GetName())
			err = client.Create(ctx, desiredResourcePtr)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		}
		return errors.WithStack(err)
	}

	// Resource exists, check if it needs to be updated.
	if !needUpdate(currentResource, desiredResource) {
		logger.Info(fmt.Sprintf("%s - up to date", reconcilerName), "namespace", desiredResourcePtr.GetNamespace(), "name", desiredResourcePtr.GetName())
		return nil
	}

	// Update the resource.
	logger.Info(fmt.Sprintf("%s - updating", reconcilerName), "namespace", desiredResourcePtr.GetNamespace(), "name", desiredResourcePtr.GetName())
	err = client.Update(ctx, desiredResourcePtr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
