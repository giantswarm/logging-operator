[![CircleCI](https://circleci.com/gh/giantswarm/logging-operator.svg?&style=shield)](https://circleci.com/gh/giantswarm/logging-operator)
[![Docker Repository on Quay](https://quay.io/repository/giantswarm/logging-operator/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/logging-operator)

# logging-operator

This operator is in charge of handling the setup and configuration of the logging stack at Giant Swarm.

It does reconcile `cluster.cluster.x-k8s.io` object and make sure each Cluster is provided with a [`promtail`](https://github.com/giantswarm/promtail-app/) instance and the necessary configuration to ship Log into [`loki`](https://github.com/giantswarm/loki-app).

## Getting started

Get the code and built it via :

```bash
git clone https://github.com/giantswarm/logging-operator.git
cd logging-operator
make
```

See `make help` for help.

## Credits

This operator was built using [`kubeduiler`](https://book.kubebuilder.io/quick-start.html).
