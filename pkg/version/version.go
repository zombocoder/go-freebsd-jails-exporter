package version

import "runtime"

var (
	Version   = "0.3.0"
	Revision  = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)
