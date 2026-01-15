[![CircleCI](https://dl.circleci.com/status-badge/img/gh/giantswarm/logging-operator/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/giantswarm/logging-operator/tree/main)

# logging-operator

## ⚠️ DEPRECATED

**This project is deprecated and no longer maintained.** It has been superseded by functionality integrated into the observability-operator.

---

This operator is in charge of handling the setup and configuration of the logging stack at Giant Swarm.

It reconciles `cluster.cluster.x-k8s.io` objects and makes sure each `Cluster` is provided with its alloy agents:
- `alloy-logs` for log collection
- `alloy-events` for kubernetes events and tracing support
and the necessary configuration to ship logs into [`loki`](https://github.com/giantswarm/loki-app).

## Getting started

Get the code and build it via:

```bash
git clone https://github.com/giantswarm/logging-operator.git
cd logging-operator
make
```

See `make help` for help.

### Generating test files

Golden test files are re-generated via:
```
make update-golden-files
```

## Architecture

The operator is built around a central reconciler, that calls multiple sub-reconcilers sequentially.
![image](images/logging-operator-architecture.png)

* Logging-Credentials are created if they don't exist. Then, their data (credentials) is used to create the next resources.
* grafana datasource configures Grafana to read data from Loki
* loki-multi-tenant-auth secures all Loki communication (per-tenant read/write access)
* logging-secret configures write access to Loki for alloy-logs
* logging-config setups some alloy-logs settings (like which logs to collect)
* logging-wiring ensures alloy-logs reads configs from previous steps
* events-logger-config setups some alloy-events settings like the `river` configuration to scrape Kubernetes Events
* events-logger-secret setups logging write credentials to access loki into the `river` configuration
* Logging agents toggle enables/disables logging agents deployment on WCs

## Gathering logs from WCs

When the need to gather logs from the WCs appears, the logging-operator will deploy alloys on those so that one may see the logs from the MC's grafana. It deploys also alloy-events to be able to scrape Kubernetes Events.
In order to achieve that, one has to label the cluster(s) one wants to gather logs from thanks to the following command :
```
kubectl label cluster -n <wc_namespace> <wc_name> giantswarm.io/logging=true
```

## Credits

This operator was built using [`kubebuilder`](https://book.kubebuilder.io/quick-start.html).
