[![CircleCI](https://dl.circleci.com/status-badge/img/gh/giantswarm/logging-operator/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/giantswarm/logging-operator/tree/main)

# logging-operator

This operator is in charge of handling the setup and configuration of the logging stack at Giant Swarm.

It reconciles `cluster.cluster.x-k8s.io` objects and makes sure each `Cluster` is provided with a [`promtail`](https://github.com/giantswarm/promtail-app) instance and the necessary configuration to ship logs into [`loki`](https://github.com/giantswarm/loki-app).

## Getting started

Get the code and build it via:

```bash
git clone https://github.com/giantswarm/logging-operator.git
cd logging-operator
make
```

See `make help` for help.

## Credits

This operator was built using [`kubebuilder`](https://book.kubebuilder.io/quick-start.html).
