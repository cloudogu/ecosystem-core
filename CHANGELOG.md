# ecosystem-core Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Changed
- Update minio to reduce CVEs with a kubectl image change.

## [v2.0.1] - 2025-12-02
### Fixed
- [#34] Update k8s-ces-gateway to v2.0.1 to use the correct controller class name `ingress-nginx`.

## [v2.0.0] - 2025-11-28
### Changed
- [#30] Updated k8s-ces-gateway, k8s-ces-assets and k8s-service-discovery to fix a redirection bug for the default dogu.

### BREAKING CHANGES

- The updated components require a migration of the existing ingressclass.
- Since it is now contained by another component (k8s-ces-gateway), the existing ingressclass has to be patched.
  - See: https://github.com/cloudogu/k8s-ces-gateway/commit/8f2d639390f5bd940866431639ef1b2f4e5d2fa7
- On new installations, this change does not require any action.

### Removed
- [#32] Removed snapshot controller and api components.

## [v1.2.0] - 2025-11-13

### Changed
- Update k8s-ces-gateway to v1.0.4. This removes the default ingress class and prevents conflicts with other ingress controllers.

### Fixed
- [#27] Update blueprint operator to v3.0.2 to fix an issue where non-referenced config entries were always empty.

## [v1.1.1] - 2025-11-12
### Changed
- [#25] Update k8s-blueprint-operator to v3.0.1

## [v1.1.0] - 2025-11-11
### Changed
- [#18] Allow global property at root level. With this change, other charts can use this chart as a dependency.
- [#19] Update components to the newest versions

## [v1.0.0] - 2025-11-07
### Changed
- [#22] update component operator dependency

## [v0.5.0] - 2025-11-07
[WARNING, THIS VERSION IS BROKEN, WE'D RECOMMEND USING THE NEXT RELEASE!]
### Added
- [#20] support config map references for components
### Changed
- [#15] update go dependencies

## [v0.4.0] - 2025-10-01
### Changed
- Update k8s-ces-gateway to v1.0.3

## [v0.3.0] - 2025-10-01
### Changed
- [#13] Update to latest component versions

## [v0.2.2] - 2025-09-30
### Fixed
- [#10] Don't overwrite external certificates by setting the right certificate type in global config

## [v0.2.1] - 2025-09-29
### Fixed
- Build and Push for release

## [v0.2.0] - 2025-09-29
### Added
- [#07] Configure annotations for loadbalancer via values.yaml

## [v0.0.1] - 2025-09-04
### Added
- [#01] Provide helm chart to install core components of the ecosystem
- [#05] post-install job to initialize default configuration