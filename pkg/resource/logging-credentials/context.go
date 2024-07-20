package loggingcredentials

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type loggingCredentialsContextKey int

var loggingCredentialsSecretKey loggingCredentialsContextKey

func NewContext(ctx context.Context, lc loggedcluster.Interface, client client.Client) (context.Context, ctrl.Result, error) {
	var loggingCredentialsSecret v1.Secret

	err := client.Get(ctx, types.NamespacedName{Name: LoggingCredentialsSecretMeta(lc).Name, Namespace: LoggingCredentialsSecretMeta(lc).Namespace},
		&loggingCredentialsSecret)
	if err != nil {
		return nil, ctrl.Result{}, err
	}

	ctx = context.WithValue(ctx, loggingCredentialsSecretKey, loggingCredentialsSecret)

	return ctx, ctrl.Result{}, nil
}

func FromContext(ctx context.Context) (v1.Secret, bool) {
	loggingCredentialsSecret, ok := ctx.Value(loggingCredentialsSecretKey).(v1.Secret)
	return loggingCredentialsSecret, ok
}
