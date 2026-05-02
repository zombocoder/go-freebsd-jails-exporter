# go-freebsd-jails-exporter

Prometheus exporter for FreeBSD jails using native cgo bindings. No shell commands are executed.

All metrics include a `host` label (system hostname) so jails can be distinguished across multiple hosts in Prometheus/Grafana.

## Requirements

- FreeBSD 13+ (optimized for 14.x)
- Go 1.22+
- CGO_ENABLED=1
- Root privileges (required for full process visibility)
- `kern.racct.enable=1` in `/boot/loader.conf` (for RACCT metrics, requires reboot)

## Build

```bash
CGO_ENABLED=1 GOOS=freebsd GOARCH=amd64 go build -o go-freebsd-jails-exporter ./cmd/go-freebsd-jails-exporter/
```

## Usage

```bash
doas ./go-freebsd-jails-exporter --web.listen-address=":9756"
```

The exporter must run as root (or via `doas`/`sudo`) to read process information from all jails. Without root, `security.bsd.see_other_uids=0` prevents visibility into jail processes.

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--web.listen-address` | `:9756` | Listen address |
| `--web.telemetry-path` | `/metrics` | Metrics path |
| `--log.level` | `info` | Log level (debug, info, warn, error) |
| `--log.format` | `logfmt` | Log format (logfmt, json) |
| `--collector.jails.enabled` | `true` | Enable jails collector |
| `--collector.processes.enabled` | `true` | Enable processes collector |
| `--collector.racct.enabled` | `true` | Enable RACCT collector |
| `--collector.timeout` | `5s` | Collector scrape timeout |

### Endpoints

| Path | Description |
|------|-------------|
| `/` | Landing page |
| `/metrics` | Prometheus metrics |
| `/healthz` | Health check |
| `/readyz` | Readiness check |

## Labels

All jail metrics include the following base labels:

| Label | Description |
|-------|-------------|
| `host` | System hostname (e.g. `bsdlab`) |
| `jid` | Jail ID |
| `name` | Jail name |

This allows grouping and filtering by host when running the exporter across multiple FreeBSD machines.

## Collectors

### Jails

Uses [`go-freebsd-jail`](https://github.com/zombocoder/go-freebsd-jail) library (native `jail_get`/`jailparam_*` APIs) for jail discovery, plus direct cgo calls for extra parameters (`parent`, `osrelease`, `osreldate`).

| Metric | Type | Extra Labels | Description |
|--------|------|--------------|-------------|
| `freebsd_jail_info` | gauge | `hostname`, `path`, `osrelease`, `osreldate` | Jail metadata |
| `freebsd_jail_state` | gauge | `state` | Jail state (active/dying) |
| `freebsd_jail_parent_jid` | gauge | | Parent jail ID |
| `freebsd_jail_persist` | gauge | | Persist flag (0/1) |
| `freebsd_jail_vnet` | gauge | | VNET enabled (0/1) |
| `freebsd_jail_ip4_addresses` | gauge | | IPv4 address count |
| `freebsd_jail_ip6_addresses` | gauge | | IPv6 address count |

### Processes

Uses `sysctl(KERN_PROC_ALL)` via cgo to count processes and threads per jail, grouped by state.

| Metric | Type | Extra Labels | Description |
|--------|------|--------------|-------------|
| `freebsd_jail_processes` | gauge | | Process count |
| `freebsd_jail_threads` | gauge | | Thread count |
| `freebsd_jail_processes_by_state` | gauge | `state` | Processes by state (running/sleeping/stopped/zombie/unknown) |

### RACCT

Uses `rctl_get_racct(2)` via cgo. Silently skipped if RACCT is not enabled in the kernel.

To enable RACCT, add to `/boot/loader.conf` and reboot:

```
kern.racct.enable=1
```

| Metric | Type | Description |
|--------|------|-------------|
| `freebsd_jail_racct_memory_used_bytes` | gauge | Memory usage |
| `freebsd_jail_racct_memory_limit_bytes` | gauge | Memory limit |
| `freebsd_jail_racct_vmemory_used_bytes` | gauge | Virtual memory usage |
| `freebsd_jail_racct_swap_used_bytes` | gauge | Swap usage |
| `freebsd_jail_racct_processes` | gauge | Process count |
| `freebsd_jail_racct_process_limit` | gauge | Process limit |
| `freebsd_jail_racct_threads` | gauge | Thread count |
| `freebsd_jail_racct_open_files` | gauge | Open file descriptors |
| `freebsd_jail_racct_cpu_percent` | gauge | CPU usage percentage |
| `freebsd_jail_racct_read_bytes_per_second` | gauge | Read throughput |
| `freebsd_jail_racct_write_bytes_per_second` | gauge | Write throughput |
| `freebsd_jail_racct_read_iops` | gauge | Read IOPS |
| `freebsd_jail_racct_write_iops` | gauge | Write IOPS |

### Exporter

| Metric | Type | Description |
|--------|------|-------------|
| `freebsd_jails_exporter_build_info` | gauge | Build info (version, revision, goversion) |
| `freebsd_jails_exporter_scrape_duration_seconds` | gauge | Scrape duration per collector |
| `freebsd_jails_exporter_scrape_errors_total` | counter | Scrape errors per collector |
| `freebsd_jails_exporter_collector_enabled` | gauge | Whether a collector is enabled |

## rc.d Service

```bash
cp scripts/freebsd/rc.d/go_freebsd_jails_exporter /usr/local/etc/rc.d/
chmod +x /usr/local/etc/rc.d/go_freebsd_jails_exporter
cp go-freebsd-jails-exporter /usr/local/bin/

# Add to /etc/rc.conf:
# go_freebsd_jails_exporter_enable="YES"
# go_freebsd_jails_exporter_user="root"

service go_freebsd_jails_exporter start
```

## Testing

```bash
CGO_ENABLED=1 go test ./...
```

## License

BSD-2-Clause
