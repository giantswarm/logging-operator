# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/giantswarm/logging-operator/compare/v0.1.2...HEAD
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
