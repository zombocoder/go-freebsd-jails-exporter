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
| `--collector.zfs.enabled` | `true` | Enable ZFS collector |
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
| `freebsd_jail_racct_cpu_time_seconds` | gauge | CPU time consumed |
| `freebsd_jail_racct_cpu_percent` | gauge | CPU usage percentage |
| `freebsd_jail_racct_memory_used_bytes` | gauge | Memory usage |
| `freebsd_jail_racct_memory_locked_bytes` | gauge | Locked memory |
| `freebsd_jail_racct_vmemory_used_bytes` | gauge | Virtual memory usage |
| `freebsd_jail_racct_swap_used_bytes` | gauge | Swap usage |
| `freebsd_jail_racct_processes` | gauge | Process count |
| `freebsd_jail_racct_threads` | gauge | Thread count |
| `freebsd_jail_racct_open_files` | gauge | Open file descriptors |
| `freebsd_jail_racct_datasize_bytes` | gauge | Data segment size |
| `freebsd_jail_racct_stacksize_bytes` | gauge | Stack size |
| `freebsd_jail_racct_pseudoterminals` | gauge | Pseudo-terminals |
| `freebsd_jail_racct_wallclock_seconds` | gauge | Wall clock time |
| `freebsd_jail_racct_read_bytes_per_second` | gauge | Read throughput |
| `freebsd_jail_racct_write_bytes_per_second` | gauge | Write throughput |
| `freebsd_jail_racct_read_iops` | gauge | Read IOPS |
| `freebsd_jail_racct_write_iops` | gauge | Write IOPS |

### ZFS

Uses [`go-freebsd-libzfs`](https://github.com/zombocoder/go-freebsd-libzfs) library to read ZFS dataset properties for each jail's filesystem path. Automatically maps jail paths to ZFS datasets.

| Metric | Type | Extra Labels | Description |
|--------|------|--------------|-------------|
| `freebsd_jail_zfs_used_bytes` | gauge | `dataset` | Space used by the dataset |
| `freebsd_jail_zfs_available_bytes` | gauge | `dataset` | Space available to the dataset |
| `freebsd_jail_zfs_referenced_bytes` | gauge | `dataset` | Space referenced by the dataset |
| `freebsd_jail_zfs_logical_used_bytes` | gauge | `dataset` | Logical space used |
| `freebsd_jail_zfs_quota_bytes` | gauge | `dataset` | Dataset quota (0 = none) |
| `freebsd_jail_zfs_refquota_bytes` | gauge | `dataset` | Reference quota (0 = none) |
| `freebsd_jail_zfs_reservation_bytes` | gauge | `dataset` | Reservation (0 = none) |
| `freebsd_jail_zfs_refreservation_bytes` | gauge | `dataset` | Reference reservation (0 = none) |
| `freebsd_jail_zfs_compressratio` | gauge | `dataset` | Compression ratio (1.0 = none) |
| `freebsd_jail_zfs_used_by_snapshots_bytes` | gauge | `dataset` | Space used by snapshots |
| `freebsd_jail_zfs_used_by_dataset_bytes` | gauge | `dataset` | Space used by the dataset itself |
| `freebsd_jail_zfs_used_by_children_bytes` | gauge | `dataset` | Space used by children |

### Exporter

| Metric | Type | Description |
|--------|------|-------------|
| `freebsd_jails_exporter_build_info` | gauge | Build info (version, revision, goversion) |
| `freebsd_jails_exporter_scrape_duration_seconds` | gauge | Scrape duration per collector |
| `freebsd_jails_exporter_scrape_errors_total` | counter | Scrape errors per collector |
| `freebsd_jails_exporter_collector_enabled` | gauge | Whether a collector is enabled |

## Grafana Dashboard

Import `grafana/dashboard.json` into Grafana. The dashboard includes:

- Overview stats (total jails, active/dying, processes, threads, memory)
- Jail inventory table
- Memory, virtual memory, and swap usage graphs
- Process and thread counts with state breakdown
- I/O throughput and IOPS
- RACCT resources (open files, CPU usage)
- ZFS usage, available space, referenced, compression ratio, and quota vs used
- Exporter health (scrape duration, errors)

All panels support filtering by Host, Jail, and Process State variables.

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

BSD-2-Clause. See [LICENSE](LICENSE).
