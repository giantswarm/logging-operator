package capicluster

import (

	"sigs.k8s.io/controller-runtime/pkg/client"

	loggedcluster "github.com/giantswarm/logging-operator/pkg/logged-cluster"
)

type Object struct {
	client.Object
	*loggedcluster.LoggingAgent
}

func (o Object) GetObject() client.Object {
	return o.Object
}

