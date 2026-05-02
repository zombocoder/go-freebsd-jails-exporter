package version

import "runtime"

var (
	Version   = "0.1.0"
	Revision  = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)
