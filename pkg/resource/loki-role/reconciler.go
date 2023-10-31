package lokirole

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/giantswarm/logging-operator/pkg/common"
	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

// Reconciler implements a reconciler.Interface to handle
// Loki role creation on CAPA MCs.
type Reconciler struct {
	client.Client
}

// ReconcileCreate ensures that the role used by Loki on CAPA MCs exists.
func (r *Reconciler) ReconcileCreate(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("lokirole create")

	if common.IsWorkloadCluster(lc) {
		logger.Info("workload cluster, skipping")
		return ctrl.Result{}, nil
	}

	roleToAssume, err := r.getRoleToAssume(ctx, lc)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	iamClient, err := r.createIamClient(ctx, roleToAssume, lc.GetRegion())
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	parsed, err := arn.Parse(roleToAssume)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	service := NewIamService(parsed.AccountID, lc.GetCloudDomain(), iamClient, logger)
	err = service.ConfigureRole(ctx, lc)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	logger.Info("lokirole - done")
	return ctrl.Result{}, nil
}

// ReconcileDelete - Deleting a role
func (r *Reconciler) ReconcileDelete(ctx context.Context, lc loggedcluster.Interface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("lokirole delete")

	if common.IsWorkloadCluster(lc) {
		logger.Info("workload cluster, skipping")
		return ctrl.Result{}, nil
	}

	roleToAssume, err := r.getRoleToAssume(ctx, lc)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	iamClient, err := r.createIamClient(ctx, roleToAssume, lc.GetRegion())
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	parsed, err := arn.Parse(roleToAssume)
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	service := NewIamService(parsed.AccountID, lc.GetCloudDomain(), iamClient, logger)
	err = service.DeleteRole(ctx, getRoleName(lc))
	if err != nil {
		return ctrl.Result{}, errors.WithStack(err)
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) getRoleToAssume(ctx context.Context, lc loggedcluster.Interface) (string, error) {
	logger := log.FromContext(ctx)

	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AWSCluster",
		Version: "v1beta2",
	})

	err := r.Client.Get(ctx, client.ObjectKey{
		Name:      lc.GetInstallationName(),
		Namespace: "org-giantswarm",
	}, cluster)
	if err != nil {
		logger.Error(err, "Missing management cluster AWSCluster CR")
		return "", errors.WithStack(err)
	}

	clusterIdentityName, found, err := unstructured.NestedString(cluster.Object, "spec", "identityRef", "name")
	if err != nil {
		logger.Error(err, "Identity name is not a string")
		return "", errors.WithStack(err)
	}
	if !found || clusterIdentityName == "" {
		logger.Info("Missing identity, skipping")
		return "", errors.New("missing management cluster identify")
	}

	clusterIdentity := &unstructured.Unstructured{}
	clusterIdentity.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AWSClusterRoleIdentity",
		Version: "v1beta2",
	})

	err = r.Client.Get(ctx, client.ObjectKey{
		Name:      clusterIdentityName,
		Namespace: cluster.GetNamespace(),
	}, clusterIdentity)
	if err != nil {
		logger.Error(err, "Missing management cluster identity AWSClusterRoleIdentity CR")
		return "", errors.WithStack(err)
	}

	roleArn, found, err := unstructured.NestedString(clusterIdentity.Object, "spec", "roleARN")
	if err != nil {
		logger.Error(err, "Role arn is not a string")
		return "", errors.WithStack(err)
	}
	if !found {
		return "", errors.New("missing role arn")
	}
	return roleArn, nil
}
