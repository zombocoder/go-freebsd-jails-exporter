# Security Policy

## Supported versions

Security fixes are made on the `main` branch and released in a new version.
Before 1.0, only the latest release receives fixes.

## Reporting a vulnerability

Please **do not open a public issue** for security problems.

Report them privately through GitHub:
[Report a vulnerability](https://github.com/zombocoder/go-freebsd-jails-exporter/security/advisories/new)
(Security tab → "Report a vulnerability").

Include what you can of:

- the exporter version (`go-freebsd-jails-exporter --version`) or commit;
- FreeBSD version (`freebsd-version -ku`);
- steps to reproduce the problem;
- the impact you expect (information disclosure, crash, memory corruption, ...).

You should get a first response within a week. Once a fix is available, the
advisory is published with credit to the reporter unless you prefer otherwise.

## Scope

The exporter usually runs as root on the jail host, so of particular interest are:

- memory-safety bugs in the cgo layer (`internal/freebsd`), such as
  use-after-free or out-of-bounds reads of sysctl or RACCT buffers;
- crashes or excessive resource use that a process inside a jail can trigger
  (for example through process names, jail parameters or dataset properties);
- information exposed over HTTP beyond the documented metrics.

### Deployment notes

The exporter has no authentication or TLS of its own. Metrics reveal jail
names, paths, hostnames and resource usage, so do not expose the listen port to
untrusted networks: bind it to a management interface
(`--web.listen-address=10.0.0.1:9756`), filter it with `pf`, or put a reverse
proxy with TLS and authentication in front of it.

Bugs in FreeBSD itself should be reported to the
[FreeBSD security team](https://www.freebsd.org/security/).
