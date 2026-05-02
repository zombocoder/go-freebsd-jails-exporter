//go:build freebsd && cgo

package collector

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector is the interface all jail metric collectors must implement.
type Collector interface {
	Name() string
	Describe(ch chan<- *prometheus.Desc)
	Collect(ctx context.Context, ch chan<- prometheus.Metric) error
}

// Registry wraps multiple Collectors and implements prometheus.Collector.
type Registry struct {
	collectors []Collector
	logger     *slog.Logger
	timeout    time.Duration

	scrapeDuration *prometheus.Desc
	scrapeErrors   *prometheus.CounterVec
	collectorUp    *prometheus.Desc
}

func NewRegistry(logger *slog.Logger, timeout time.Duration, collectors ...Collector) *Registry {
	scrapeErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "freebsd_jails_exporter",
			Name:      "scrape_errors_total",
			Help:      "Total number of scrape errors by collector.",
		},
		[]string{"collector"},
	)

	// Pre-initialize error counters so they always appear in output.
	for _, c := range collectors {
		scrapeErrors.WithLabelValues(c.Name())
	}

	return &Registry{
		collectors: collectors,
		logger:     logger,
		timeout:    timeout,
		scrapeDuration: prometheus.NewDesc(
			"freebsd_jails_exporter_scrape_duration_seconds",
			"Duration of a collector scrape.",
			[]string{"collector"},
			nil,
		),
		scrapeErrors: scrapeErrors,
		collectorUp: prometheus.NewDesc(
			"freebsd_jails_exporter_collector_enabled",
			"Whether a collector is enabled.",
			[]string{"collector"},
			nil,
		),
	}
}

func (r *Registry) Describe(ch chan<- *prometheus.Desc) {
	ch <- r.scrapeDuration
	r.scrapeErrors.Describe(ch)
	ch <- r.collectorUp
	for _, c := range r.collectors {
		c.Describe(ch)
	}
}

func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup

	for _, c := range r.collectors {
		ch <- prometheus.MustNewConstMetric(r.collectorUp, prometheus.GaugeValue, 1, c.Name())

		wg.Add(1)
		go func(c Collector) {
			defer wg.Done()
			r.collect(c, ch)
		}(c)
	}

	wg.Wait()
	r.scrapeErrors.Collect(ch)
}

func (r *Registry) collect(c Collector, ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	start := time.Now()
	err := c.Collect(ctx, ch)
	duration := time.Since(start).Seconds()

	ch <- prometheus.MustNewConstMetric(r.scrapeDuration, prometheus.GaugeValue, duration, c.Name())

	if err != nil {
		r.scrapeErrors.WithLabelValues(c.Name()).Inc()
		r.logger.Error("collector failed",
			"collector", c.Name(),
			"duration", fmt.Sprintf("%.4fs", duration),
			"error", err,
		)
	} else {
		r.logger.Debug("collector succeeded",
			"collector", c.Name(),
			"duration", fmt.Sprintf("%.4fs", duration),
		)
	}
}
