[![CircleCI](https://dl.circleci.com/status-badge/img/gh/giantswarm/logging-operator/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/giantswarm/logging-operator/tree/main)

# logging-operator

This operator is in charge of handling the setup and configuration of the logging stack at Giant Swarm.

It reconciles `cluster.cluster.x-k8s.io` objects and makes sure each `Cluster` is provided with logging agents:
- [`promtail`](https://github.com/giantswarm/promtail-app) instance
- [`grafana-agent`](https://github.com/giantswarm/grafana-agent-app) instance
and the necessary configuration to ship logs into [`loki`](https://github.com/giantswarm/loki-app).

## Getting started

Get the code and build it via:

```bash
git clone https://github.com/giantswarm/logging-operator.git
cd logging-operator
make
```

See `make help` for help.

## Architecture

The operator is built around a central reconciler, that calls multiple sub-reconcilers sequentially.
![image](images/logging-operator-architecture.png)

* Logging-Credentials are created if they don't exist. Then, their data (credentials) is used to create the next resources.
* grafana datasource configures Grafana to read data from Loki
* loki-multi-tenant-auth secures all Loki communication (per-tenant read/write access)
* promtail client configures write access to Loki for Promtail
* Promtail config setups some Promtail settings (like which logs to collect)
* Promtail wiring ensures promtail-app reads configs from previous steps
* grafana-agent config setups some grafana-agent settings like the `river` configuration to scrape Kubernetes Events
* grafana-agent secret setups logging write credentials to access loki into the `river` configuration
* Logging agents toggle enables/disables logging agents deployment on WCs

## Gathering logs from WCs

When the need to gather logs from the WCs appears, the logging-operator will deploy promtail on those so that one may see the logs from the MC's grafana. It deploys also grafana-agent to be able to scrape Kubernetes Events.
In order to achieve that, one has to label the cluster(s) one wants to gather logs from thanks to the following command :
```
kubectl label cluster -n <wc_namespace> <wc_name> giantswarm.io/logging=true
```

## Credits

This operator was built using [`kubebuilder`](https://book.kubebuilder.io/quick-start.html).
