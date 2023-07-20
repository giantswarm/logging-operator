# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Push app to aws-app-catalog
- Commented reconcilers creation for Vintage WC and CAPI clusters - not supported yet.

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

[Unreleased]: https://github.com/giantswarm/logging-operator/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/giantswarm/logging-operator/releases/tag/v0.0.1
