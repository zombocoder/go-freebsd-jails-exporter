package version

import "runtime"

var (
	Version   = "0.3.1"
	Revision  = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)
