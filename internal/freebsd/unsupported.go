//go:build !freebsd || !cgo

package freebsd

import "errors"

// errUnsupported is returned by every function in this package when the
// exporter is built for a platform other than FreeBSD, or without cgo.
var errUnsupported = errors.New("freebsd: not supported on this platform (requires FreeBSD and cgo)")
