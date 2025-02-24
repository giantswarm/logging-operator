# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Add `grafana-organization-reconciler` to be used for tenant governance.
- Add tenant filtering in Alloy config

### Fixed

- Fix non-working log lines dropping on missing tenant id

## [0.22.0] - 2025-02-12

### Changed

- Collect cluster wide events for Nodes in the default namespace.

## [0.21.0] - 2025-02-03

### Added

- Add namespace, pod, and container log label on logs coming from PodLogs

### Fixed

- Fix the `job` label in logs, by removing the associated relabeling rule.

## [0.20.1] - 2025-01-30


### Fixed

- Fix invalid workload cluster pod logs selectors.

## [0.20.0] - 2025-01-14

### Added

- Add support for customer log tenancy via pod logs.

## [0.19.0] - 2025-01-07

### Removed

- Remove Loki datasource generation on CAPI.

## [0.18.0] - 2025-01-07

### Changed

- Replace multi-tenant-proxy with ingress auth map on CAPI.

## [0.17.0] - 2025-01-07

### Changed

- Use `giantswarm` tenant by default instead of cluster name.

## [0.16.0] - 2024-11-26

### Added

- Add kubernetes events logging in Alloy.
- Add support for Private CAs in alloy logs.
- Add KubeEventsLogger option and related methods in loggedCLuster package.
- Add `events-logger` flag in the operator.
- Add toggle for `events-logger` in observability-bundle configmap.
- Add tests for `alloy-events` in events-logger-config.

### Changed

- Disable grafana agent usage data reporting.
- Move Grafana-Agent config to a template instead of go structs.

### Fixed

- Fix logging-config unit tests

## [0.15.2] - 2024-11-13

### Added

- Enable VPA on alloy if the deployed alloy version supports it.

## [0.15.1] - 2024-11-04

### Fixed

- Fixes the current version that enabled alloy-logs as the new secret mechanism only works with alloy 0.4.0 which is is the observability bundle 1.6.0

## [0.15.0] - 2024-10-31

### Changed

- Reconcile clusters when the observability bundle version changes.

### Fixed

- Disable crd installation from grafana agent as this is causing issues with the new v29 releases.

## [0.14.0] - 2024-10-29

### Changed

- Change default logging agent to Alloy instead of Promtail.

## [0.13.0] - 2024-10-24

### Added

- Add manual e2e testing procedure and script.
- [Alloy] Add capability to dynamically configure log targets using `PodLogs` [#3618](https://github.com/giantswarm/roadmap/issues/3518)
  - Changes Alloy config to use PodLogs for Kubernetes pods discovery.
  - There is a performance impact on Kubernetes API
  - Available from observability-bundle v17.0.0

### Changed

- [Alloy] Enable clustering
- Expose healthcheck port for kube-linter.

### Removed

- Remove the alloy-secret resource which is no longer needed as Alloy secret was moved into logging-secret

### Fixed

- Fix circleci config.

## [0.12.1] - 2024-09-23

### Fixed

- Fix usage of structured metadata for clusters before v20.

## [0.12.0] - 2024-09-23

### Changed

- Move high cardinality values into structured-metadata:
  - kubernetes audit log `resource` label
  - `filename` label
  - log `output stream` label
- Rename `node_name` label into `node` to match metric label.

## [0.11.2] - 2024-09-18

### Fixed

- Fix v0.11.1 release was not published

## [0.11.1] - 2024-09-18

### Fixed

- Fix v0.11.0 release was not published

## [0.11.0] - 2024-09-18

### Fixed

- Fix Alloy logs secret handling

## [0.10.0] - 2024-09-10

### Changed

- Use grafana multi-tenant-proxy public types instead of vendoring them.

### Fixed

- Disable logger development mode to avoid panicking, use zap as logger.

### Removed

- Delete old loki-auth-proxy configuration in favor of the new grafana-auth-proxy.
- Remove old loki-auth reconciler.

## [0.9.0] - 2024-09-03

### Changed

- Change the datasource url to be the multi-tenant-proxy in front of the loki-gateway.
- Add secret management for the proxy by duplicating loki-auth.

### Fixed

- Fix incorrect alloy security context.

## [0.8.0] - 2024-08-27

### Added

- Add helm chart templating test in ci pipeline.
- Add tests with ats in ci pipeline.

## [0.7.3] - 2024-08-13

### Fixed

- Fix required observability bundle version to run Alloy as logging agent.

## [0.7.2] - 2024-08-12

### Fixed

- Fix alloy-secret naming
  - Rename secret to alloy-logging-secret
  - Prefix secret with cluster name
  - Create secret in the cluster namespace

## [0.7.1] - 2024-08-08

### Fixed

- Fix incorrect Alloy secret and Alloy config templating.
- Rename `alloy-logs` to came case `alloyLogs` in the observability-bundle config.

## [0.7.0] - 2024-07-19

### Added

- Add support for Alloy as logging agent
  Add `--logging-agent` flag to toggle between Promtail and Alloy

## [0.6.0] - 2024-07-09

### Changed

- Use a deployment for the grafana-agent instance used to collect kubernetes events to avoid using too much resources on clusters as long as we use promtail.

## [0.5.5] - 2024-06-14


### Fixed

- Fix reconciliation errors when adding or removing the finalizer on the Cluster CR.

## [0.5.4] - 2024-04-23

### Changed

- Replace promtail occurrences by logging to be more generic.

### Fixed

- Delete leftover configmaps and secrets while cluster deleting.

## [0.5.3] - 2024-04-09

### Changed

- Reduce audit log cardinality by ignoring rotated audit log files to avoid duplicate audit logs.

## [0.5.2] - 2024-03-06

### Changed

- Update deprecated `targetPort` to `port` in PodMonitor.

## [0.5.1] - 2024-02-21

### Fixed

- This feature fixes reconciliation by support requeuing of failed request that do not necessarily need to be an error (missing app due to slow cluster boostrapping).

## [0.5.0] - 2024-02-19

### Removed

- Remove multi-tenant proxy restart hack.

## [0.4.4] - 2024-01-22

### Changed

- Push to CAPV.

## [0.4.3] - 2024-01-17

### Fixed

- Ignore not found error for clusters that have logging disabled.

## [0.4.2] - 2024-01-11

### Fixed

- Watch observability-bundle apps to handle bundle migration to v1.0.0.

## [0.4.1] - 2024-01-09

### Fixed

- Fix reconcile errors (grafana-agent-config, promtail-wiring, cluster not found).

## [0.4.0] - 2024-01-04

### Added

- Expose profiles in the controller and add conditional profiling annotations in the deployment.

### Changed

- Configure `gsoci.azurecr.io` as the default container image registry.
- Drop system audit logs from Promtail's scrape target

### Fixed

- Replace systemd_unit label with syslog identifier for system logs without systemd_unit label
- Fixed podmonitor

## [0.3.1] - 2023-12-04

### Changed

- enable logging on WCs by default.
- push to CAPZ and CAPVCD collections

## [0.3.0] - 2023-11-21

### Changed

- Add labels to kubernetes audit logs to reduce rate limiting and help discovering logs.

## [0.2.2] - 2023-11-14

### Fixed

- Fix grafana-agent configMap creation on CAPI.

## [0.2.1] - 2023-11-10

### Changed

- default logging behavior on WCs reverted to disable.

## [0.2.0] - 2023-11-09

### Added

- Configure grafana-agent config to grab Kubernetes Events and send them to Loki.
- Create grafana-agent extra secret to store logging write credentials.
- Prepare some stuff for CAPI.
- Upgrade go dependencies.

### Changed

- Change default logging behavior on WCs >= 19.1.0. Logging is now enabled by default.
- Improve network policy and minor go fixes.

## [0.1.4] - 2023-10-31

### Changed

- Push to CAPA app collection.

## [0.1.3] - 2023-10-18

### Changed

- Configure correct app depending on observability-bundle version.

## [0.1.2] - 2023-10-17

### Fixed

- Revert support for observability-bundle 1.0.0.

## [0.1.1] - 2023-10-17

### Added

- Add support for observability-bundle 1.0.0.

## [0.1.0] - 2023-10-17

### Changed

- Only workload clusters release >= v19.1.0 can enable logging.
- each cluster has a dedicated user
- each cluster sends data as a different tenant
- update logging-credentials secret format

## [0.0.7] - 2023-10-03

### Added

- Audit logs in promtail config.
- Add condition for PSP installation in helm chart.

### Changed

- Logs labels updated to ease navigation.

## [0.0.6] - 2023-09-28

### Fixed

- Ensure we do not delete observability-bundle user configs for workload clusters.

## [0.0.5] - 2023-09-20

### Added

- Scrape logs from kube-system and giantswarm namespaces only for WC clusters.

## [0.0.4] - 2023-09-18

### Changed

- Adapted code to handle promtail deployment in WCs.

## [0.0.3] - 2023-07-27

### Fixed

- Add missing RBAC access to apps/deployment resources.

## [0.0.2] - 2023-07-26

### Added

- promtail-config reconciler: creates promtail-config as extra-values.

### Changed

- Push app to aws-app-catalog
- Commented reconcilers creation for Vintage WC and CAPI clusters - not supported yet.
- image tag is defined from chart version

### Fixed

- PSP permissions to update app
- fix CODEOWNERS

## [0.0.1] - 2023-07-13

### Added

- Add Helm chart
- Implement LoggingReconciler abstraction
- Implement promtrail-wiring resource
- Implement promtail-toggle resource
- Implement finalizer handling
- Add reconciler.Interface
- Add controller for Vintage Management Cluster via corev1.Service
- Add controller for Cluster API, cluster.x-k8s.io/v1beta1
- Add operator basics with kubebuilder
- Add '--vintage' toggle
- Add controller for Workload Management Cluster using cluster.x-k8s.io/v1beta1```

[Unreleased]: https://github.com/giantswarm/logging-operator/compare/v0.22.0...HEAD
[0.22.0]: https://github.com/giantswarm/logging-operator/compare/v0.21.0...v0.22.0
[0.21.0]: https://github.com/giantswarm/logging-operator/compare/v0.20.1...v0.21.0
[0.20.1]: https://github.com/giantswarm/logging-operator/compare/v0.20.0...v0.20.1
[0.20.0]: https://github.com/giantswarm/logging-operator/compare/v0.19.0...v0.20.0
[0.19.0]: https://github.com/giantswarm/logging-operator/compare/v0.18.0...v0.19.0
[0.18.0]: https://github.com/giantswarm/logging-operator/compare/v0.17.0...v0.18.0
[0.17.0]: https://github.com/giantswarm/logging-operator/compare/v0.16.0...v0.17.0
[0.16.0]: https://github.com/giantswarm/logging-operator/compare/v0.15.2...v0.16.0
[0.15.2]: https://github.com/giantswarm/logging-operator/compare/v0.15.1...v0.15.2
[0.15.1]: https://github.com/giantswarm/logging-operator/compare/v0.15.0...v0.15.1
[0.15.0]: https://github.com/giantswarm/logging-operator/compare/v0.14.0...v0.15.0
[0.14.0]: https://github.com/giantswarm/logging-operator/compare/v0.13.0...v0.14.0
[0.13.0]: https://github.com/giantswarm/logging-operator/compare/v0.12.1...v0.13.0
[0.12.1]: https://github.com/giantswarm/logging-operator/compare/v0.12.0...v0.12.1
[0.12.0]: https://github.com/giantswarm/logging-operator/compare/v0.11.2...v0.12.0
[0.11.2]: https://github.com/giantswarm/logging-operator/compare/v0.11.1...v0.11.2
[0.11.1]: https://github.com/giantswarm/logging-operator/compare/v0.11.0...v0.11.1
[0.11.0]: https://github.com/giantswarm/logging-operator/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/giantswarm/logging-operator/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/giantswarm/logging-operator/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/giantswarm/logging-operator/compare/v0.7.3...v0.8.0
[0.7.3]: https://github.com/giantswarm/logging-operator/compare/v0.7.2...v0.7.3
[0.7.2]: https://github.com/giantswarm/logging-operator/compare/v0.7.1...v0.7.2
[0.7.1]: https://github.com/giantswarm/logging-operator/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/giantswarm/logging-operator/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/giantswarm/logging-operator/compare/v0.5.5...v0.6.0
[0.5.5]: https://github.com/giantswarm/logging-operator/compare/v0.5.4...v0.5.5
[0.5.4]: https://github.com/giantswarm/logging-operator/compare/v0.5.3...v0.5.4
[0.5.3]: https://github.com/giantswarm/logging-operator/compare/v0.5.2...v0.5.3
[0.5.2]: https://github.com/giantswarm/logging-operator/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/giantswarm/logging-operator/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/giantswarm/logging-operator/compare/v0.4.4...v0.5.0
[0.4.4]: https://github.com/giantswarm/logging-operator/compare/v0.4.3...v0.4.4
[0.4.3]: https://github.com/giantswarm/logging-operator/compare/v0.4.2...v0.4.3
[0.4.2]: https://github.com/giantswarm/logging-operator/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/giantswarm/logging-operator/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/giantswarm/logging-operator/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/giantswarm/logging-operator/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/logging-operator/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/giantswarm/logging-operator/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/giantswarm/logging-operator/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/giantswarm/logging-operator/compare/v0.1.4...v0.2.0
[0.1.4]: https://github.com/giantswarm/logging-operator/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/giantswarm/logging-operator/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/giantswarm/logging-operator/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/giantswarm/logging-operator/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/giantswarm/logging-operator/compare/v0.0.7...v0.1.0
[0.0.7]: https://github.com/giantswarm/logging-operator/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/giantswarm/logging-operator/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/giantswarm/logging-operator/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/giantswarm/logging-operator/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/giantswarm/logging-operator/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/giantswarm/logging-operator/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/giantswarm/logging-operator/releases/tag/v0.0.1
