# Changelog

All notable changes to this project are documented here. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
follows [Semantic Versioning](https://semver.org/). Before 1.0.0, minor
versions may contain breaking changes to metric names or flags; they are listed
under "Changed".

## [Unreleased]

### Added

- GitHub Actions CI: gofmt, `go mod tidy`, vet and unit tests on Linux, plus
  build, tests and an end-to-end smoke test (throwaway ZFS pool and jail) in
  FreeBSD 14.5 and 15.1 VMs.
- Release workflow: pushing a `v*` tag publishes FreeBSD 14 and 15 tarballs with
  SHA-256 checksums to GitHub Releases.
- `Makefile` with `build` (embeds version, revision and build date), `check`,
  `smoke`, `install-service`, `deps` and `dist` targets.
- Community files: contributing guide, code of conduct, security policy, issue
  and PR templates, Dependabot configuration, funding links.
- README: badges, `/usr/src` build requirement, `-ldflags` build example and
  troubleshooting for `error obtaining VCS status`.

### Fixed

- `freebsd_jail_persist` was always 0: libjail exports boolean parameters as
  `true`/`false`, but the collector compared them with `1`.
- The `internal/freebsd` stubs for non-FreeBSD platforms did not compile
  (`errUnsupported` was undefined).
- gofmt formatting in `internal/freebsd`.

## [0.2.0] - 2026-05-02

### Added

- ZFS collector: per-jail dataset usage, available space, referenced and
  logical space, quotas, reservations, compression ratio and space breakdown,
  via [go-freebsd-libzfs](https://github.com/zombocoder/go-freebsd-libzfs).
- ZFS panels in the Grafana dashboard.

### Fixed

- RACCT metrics.

## [0.1.0] - 2026-05-02

### Added

- Initial release: jails collector (via
  [go-freebsd-jail](https://github.com/zombocoder/go-freebsd-jail)), processes
  collector (sysctl `KERN_PROC`) and RACCT collector (`rctl_get_racct`), with a
  `host` label on all metrics.
- Grafana dashboard, rc.d service script and unit tests.

[Unreleased]: https://github.com/zombocoder/go-freebsd-jails-exporter/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/zombocoder/go-freebsd-jails-exporter/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/zombocoder/go-freebsd-jails-exporter/releases/tag/v0.1.0
