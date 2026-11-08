## What and why

<!-- What does this change do, and what problem does it solve? Link issues with "Fixes #123". -->

## How it was tested

<!-- CI builds and runs an end-to-end smoke test on FreeBSD 14.x and 15.x VMs.
     If you also tested on a real host, say how, e.g.:
     FreeBSD 15.1, 12 jails managed by Bastille, `doas make smoke`, compared with `rctl -u` -->

## Checklist

- [ ] `make check` passes (gofmt, go mod tidy, vet, tests)
- [ ] New parsing or mapping logic has unit tests in `tests/`
- [ ] New metrics follow Prometheus [naming conventions](https://prometheus.io/docs/practices/naming/), carry the `host`, `jid` and `name` labels, and are listed in the README
- [ ] cgo code frees every C allocation it makes
- [ ] No shell commands are executed by the exporter
- [ ] Grafana dashboard updated if it should show the new data
- [ ] User-visible changes are listed under "Unreleased" in `CHANGELOG.md`
