package buildinfo

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zombocoder/go-freebsd-jails-exporter/pkg/version"
)

func NewCollector() prometheus.Collector {
	return prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "freebsd_jails_exporter",
			Name:      "build_info",
			Help:      "A metric with a constant '1' value labeled by version, revision, and goversion from which the exporter was built.",
			ConstLabels: prometheus.Labels{
				"version":   version.Version,
				"revision":  version.Revision,
				"goversion": version.GoVersion,
			},
		},
		func() float64 { return 1 },
	)
}
