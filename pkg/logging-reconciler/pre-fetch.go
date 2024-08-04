package loggingreconciler

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type PreFetcher func(context.Context, loggedcluster.Interface, client.Client) (context.Context, ctrl.Result, error)
