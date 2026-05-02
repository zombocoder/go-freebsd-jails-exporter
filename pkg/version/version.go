package version

import "runtime"

var (
	Version   = "0.2.0"
	Revision  = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)
