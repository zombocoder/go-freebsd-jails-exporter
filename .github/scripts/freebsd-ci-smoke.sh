#!/bin/sh
# End-to-end smoke test for a built exporter binary.
#
# Creates a throwaway file-backed ZFS pool and a persistent jail rooted on one
# of its datasets, runs a process inside the jail, starts the exporter and
# checks that /metrics reports the jail from every collector. Everything it
# creates is removed on exit.
#
# Must run as root on a throwaway machine (CI VM).
#
# Usage: sh freebsd-ci-smoke.sh [path/to/go-freebsd-jails-exporter]
set -eu

BIN=$(realpath "${1:-./go-freebsd-jails-exporter}")
POOL=jexci
DIR=/var/tmp/$POOL
JAIL=jexci
ROOT=/$POOL/jails/$JAIL
ADDR=127.0.0.1:19756
TMP=$(mktemp -d)
EXPORTER_PID=

cleanup() {
	set +e
	[ -n "$EXPORTER_PID" ] && kill "$EXPORTER_PID" 2>/dev/null
	jail -r "$JAIL" 2>/dev/null
	zpool destroy -f "$POOL" 2>/dev/null
	rm -rf "$DIR" "$TMP"
}
trap cleanup EXIT

fail() {
	echo "::error::smoke test: $*"
	echo "--- exporter log"
	cat "$TMP/exporter.log" 2>/dev/null
	exit 1
}

if zpool list -H "$POOL" >/dev/null 2>&1 || jls -j "$JAIL" >/dev/null 2>&1; then
	echo "pool or jail '$POOL' already exists; refusing to touch it" >&2
	trap - EXIT
	exit 1
fi

echo "--- creating pool $POOL and jail $JAIL"
kldload -n zfs
mkdir -p "$DIR"
truncate -s 256m "$DIR/disk0"
zpool create -O mountpoint="/$POOL" "$POOL" "$DIR/disk0"
zfs create -p -o compression=lz4 "$POOL/jails/$JAIL"

# A minimal jail root: /rescue is a statically linked multi-call binary that
# picks its behavior from argv[0], so a copy named "sleep" is enough.
mkdir -p "$ROOT/bin"
cp /rescue/sleep "$ROOT/bin/sleep"
jail -c name="$JAIL" path="$ROOT" host.hostname="$JAIL" persist
jexec "$JAIL" /bin/sleep 600 &
sleep 1
jls

echo "--- starting exporter"
"$BIN" --version
"$BIN" --web.listen-address="$ADDR" --log.level=debug >"$TMP/exporter.log" 2>&1 &
EXPORTER_PID=$!

i=0
until fetch -qo /dev/null "http://$ADDR/healthz" 2>/dev/null; do
	i=$((i + 1))
	[ "$i" -ge 20 ] && fail "exporter did not become healthy"
	kill -0 "$EXPORTER_PID" 2>/dev/null || fail "exporter exited"
	sleep 0.5
done

fetch -qo /dev/null "http://$ADDR/readyz" || fail "/readyz failed"
fetch -qo "$TMP/metrics" "http://$ADDR/metrics" || fail "/metrics failed"

# expect_metric <metric name> <description>
expect_metric() {
	grep -E "^$1\{[^}]*name=\"$JAIL\"" "$TMP/metrics" >/dev/null ||
		fail "missing $1 for jail $JAIL ($2)"
	echo "ok: $1"
}

expect_metric freebsd_jail_info "jails collector"
expect_metric freebsd_jail_processes "processes collector"
expect_metric freebsd_jail_zfs_used_bytes "zfs collector"

if [ "$(sysctl -n kern.racct.enable 2>/dev/null || echo 0)" = 1 ]; then
	expect_metric freebsd_jail_racct_processes "racct collector"
else
	echo "skip: RACCT is disabled (kern.racct.enable=0)"
fi

processes=$(grep -E "^freebsd_jail_processes\{[^}]*name=\"$JAIL\"" "$TMP/metrics" | awk '{print $2}')
[ "${processes%%.*}" -ge 1 ] || fail "expected at least one process in jail $JAIL, got $processes"
echo "ok: $processes process(es) in jail $JAIL"

grep -E "^freebsd_jail_persist\{[^}]*name=\"$JAIL\"[^}]*\} 1$" "$TMP/metrics" >/dev/null ||
	fail "expected freebsd_jail_persist 1 for jail $JAIL (created with persist)"
echo "ok: persist flag reported"

errors=$(grep -E '^freebsd_jails_exporter_scrape_errors_total' "$TMP/metrics" | awk '$2 != 0')
[ -z "$errors" ] || fail "collectors reported scrape errors: $errors"
echo "ok: no scrape errors"

echo "--- jail metrics"
grep -E "name=\"$JAIL\"" "$TMP/metrics" | grep -v '_by_state' | head -40
echo "smoke test passed"
