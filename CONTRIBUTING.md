# Contributing to go-freebsd-jails-exporter

Thanks for your interest in improving the exporter! Bug reports, new metrics,
dashboard improvements, fixes, tests and documentation are all welcome. If you
run jails with Bastille, Sylve, iocage, pot or plain `jail.conf`, reports about
how the exporter behaves on your setup are especially useful.

By participating you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
Security problems should be reported privately, see [SECURITY.md](SECURITY.md).

## Table of Contents

- [Getting Started](#getting-started)
- [Running Tests](#running-tests)
- [Continuous Integration](#continuous-integration)
- [Code Style](#code-style)
- [Adding a Metric or Collector](#adding-a-metric-or-collector)
- [Submitting Changes](#submitting-changes)
- [Releases](#releases)
- [Where Help Is Wanted](#where-help-is-wanted)

## Getting Started

### Prerequisites

- FreeBSD 14 or 15, amd64 or arm64 (a VM works fine)
- Go 1.24+ (`pkg install go`)
- The FreeBSD source tree, or at least the OpenZFS headers, in `/usr/src`
  (see [README: FreeBSD sources](README.md#freebsd-sources-usrsrc))
- Root (or `doas`) to run the exporter with full visibility and for `make smoke`
- Optional: `kern.racct.enable=1` in `/boot/loader.conf` for RACCT metrics

`doas sh .github/scripts/freebsd-ci-prepare.sh` installs Go and extracts just
the OpenZFS headers for the running release; it is what CI uses.

### Setup

```sh
git clone https://github.com/<your-username>/go-freebsd-jails-exporter.git
cd go-freebsd-jails-exporter
make deps     # checks Go, OpenZFS headers and RACCT
make build    # ./go-freebsd-jails-exporter with version info
doas ./go-freebsd-jails-exporter --log.level=debug
```

Then open <http://localhost:9756/metrics>.

Build as your regular user and use `doas` only to run the binary. Building as
root in a clone owned by another user fails with
`error obtaining VCS status: exit status 128` (see the README troubleshooting
section).

## Running Tests

| Command | Runs as | What it covers |
|---|---|---|
| `make check` | any user | gofmt, `go mod tidy`, `go vet` and the unit tests in `tests/`. |
| `doas make smoke` | root | End-to-end: creates a throwaway file-backed pool `jexci` under `/var/tmp` and a jail `jexci` rooted on it, starts the exporter and checks that every collector reports the jail. Everything is removed afterwards. |

`make smoke` refuses to run if a pool or jail named `jexci` already exists, but
it does create a pool and a jail as root. Prefer a VM or test machine over a
production host.

Most of `tests/` only builds on FreeBSD with cgo. On other systems the
`internal/freebsd` package compiles to stubs, so `go vet ./...` and
`gofmt` still work for editing on a laptop.

## Continuous Integration

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on every push to
`main` and every pull request:

- **Lint** (Linux): gofmt, `go mod tidy -diff`, `go vet`, platform-independent
  tests, and a JSON check of the Grafana dashboard.
- **FreeBSD 14.5 and 15.1** (QEMU VMs via
  [vmactions/freebsd-vm](https://github.com/vmactions/freebsd-vm)): installs Go
  and the OpenZFS headers, runs `go vet` and the tests, builds the binary and
  runs the [smoke test](.github/scripts/freebsd-ci-smoke.sh) as root.

RACCT is disabled in the CI VMs (enabling it needs a reboot), so RACCT metrics
are only checked by the smoke test when you run it on a host with
`kern.racct.enable=1`.

## Code Style

- Standard Go conventions; `gofmt` is enforced by CI.
- **No shell commands.** The exporter reads everything through system calls,
  sysctl and native libraries (libjail, libzfs). Shelling out is only
  acceptable in tests and CI scripts.
- **cgo:** every `C.CString`/`C.malloc` has a matching `C.free`; check C
  pointers before use; keep cgo code in `internal/freebsd/*_cgo.go` behind
  `//go:build freebsd && cgo`, with a matching stub in `*_unsupported.go`
  that returns `errUnsupported`.
- **Collectors fail softly.** Return an error from `Collect` (it is counted in
  `freebsd_jails_exporter_scrape_errors_total`) rather than panicking, and skip
  quietly when a feature is unavailable (e.g. RACCT disabled).
- Wrap errors with context: `fmt.Errorf("listing jails: %w", err)`.
- Logging uses `log/slog`; per-scrape details go to `Debug`.

## Adding a Metric or Collector

```
cmd/go-freebsd-jails-exporter/   main: flags, HTTP server, collector wiring
internal/collector/              Collector interface, registry and collectors
internal/freebsd/                cgo bindings (sysctl, processes, RACCT, jail params) + stubs
internal/buildinfo, logging, config
pkg/version/                     version info set via -ldflags
tests/                           unit tests
grafana/dashboard.json           dashboard
scripts/freebsd/rc.d/            rc.d service script
.github/                         CI, release workflow, CI scripts, templates
```

- **Metric names** start with `freebsd_jail_` (per-jail) or
  `freebsd_jails_exporter_` (about the exporter itself), use base units
  (`_bytes`, `_seconds`), and end in `_total` for counters. See the
  [Prometheus naming guide](https://prometheus.io/docs/practices/naming/).
- **Labels:** every per-jail metric has `host`, `jid` and `name`. Avoid
  labels with unbounded values (PIDs, command lines).
- **New collector:** implement `collector.Collector` (`Name`, `Describe`,
  `Collect`) in `internal/collector/`, add a `--collector.<name>.enabled` flag
  in `main.go`, and register it there.
- **Document** new metrics in the README tables, and add a dashboard panel if
  it is useful to see.
- **Test** parsing and mapping logic in `tests/`, and extend
  `.github/scripts/freebsd-ci-smoke.sh` with an `expect_metric` line so CI
  checks the metric end to end.

## Submitting Changes

1. Fork and create a branch from `main`.
2. Make the change with tests, update the README/dashboard if needed, and add an
   entry under "Unreleased" in [CHANGELOG.md](CHANGELOG.md).
3. Run `make check` (and `doas make smoke` if you have a test machine) on FreeBSD.
4. Open a pull request; the template has a short checklist. CI must be green.

Commit messages: a short `type: summary` line (e.g.
`feat: add network interface metrics`, `fix: racct cpu percent overflow`), a
blank line, then what changed and why.

## Releases

Maintainers release by moving the "Unreleased" changelog entries under a new
version heading and pushing a tag:

```sh
git tag -a v0.3.0 -m "v0.3.0"
git push origin v0.3.0
```

The [release workflow](.github/workflows/release.yml) builds tarballs on
FreeBSD 14 and 15 (binary, rc.d script, Grafana dashboard, docs, SHA-256
checksum), runs the smoke test on each, and publishes a GitHub release with the
notes from `CHANGELOG.md`.

## Where Help Is Wanted

- **Network metrics** per jail (VNET interfaces, epairs)
- **Packaging:** a FreeBSD port / `pkg` package
- **Dashboard** improvements and alerting rule examples
- **Testing** with different jail managers and on arm64
- **Hierarchical jails** (child jails) coverage

Questions and design discussions are welcome in GitHub issues.
