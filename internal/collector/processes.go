//go:build freebsd && cgo

package collector

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zombocoder/go-freebsd-jail/jail"
	"github.com/zombocoder/go-freebsd-jails-exporter/internal/freebsd"
)

type ProcessesCollector struct {
	logger *slog.Logger
	host   string

	processes      *prometheus.Desc
	threads        *prometheus.Desc
	processByState *prometheus.Desc
}

func NewProcessesCollector(logger *slog.Logger, host string) *ProcessesCollector {
	labels := []string{"host", "jid", "name"}

	return &ProcessesCollector{
		logger: logger,
		host:   host,
		processes: prometheus.NewDesc(
			"freebsd_jail_processes",
			"Number of processes in the jail.",
			labels, nil,
		),
		threads: prometheus.NewDesc(
			"freebsd_jail_threads",
			"Number of threads in the jail.",
			labels, nil,
		),
		processByState: prometheus.NewDesc(
			"freebsd_jail_processes_by_state",
			"Number of processes in the jail by state.",
			[]string{"host", "jid", "name", "state"}, nil,
		),
	}
}

func (c *ProcessesCollector) Name() string { return "processes" }

func (c *ProcessesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.processes
	ch <- c.threads
	ch <- c.processByState
}

func (c *ProcessesCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	jails, err := jail.List()
	if err != nil {
		return fmt.Errorf("listing jails: %w", err)
	}

	allStats, err := freebsd.GetAllJailProcStats()
	if err != nil {
		return fmt.Errorf("getting process stats: %w", err)
	}

	for _, j := range jails {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		jid := int(j.JID)
		jidStr := strconv.Itoa(jid)
		name := j.Name

		stats, ok := allStats[jid]
		if !ok {
			stats = &freebsd.JailProcStats{ByState: make(map[freebsd.ProcState]int)}
		}

		ch <- prometheus.MustNewConstMetric(c.processes, prometheus.GaugeValue,
			float64(stats.Processes), c.host, jidStr, name)

		ch <- prometheus.MustNewConstMetric(c.threads, prometheus.GaugeValue,
			float64(stats.Threads), c.host, jidStr, name)

		// Emit all known states, even if zero.
		for _, state := range []freebsd.ProcState{
			freebsd.ProcStateRunning,
			freebsd.ProcStateSleeping,
			freebsd.ProcStateStopped,
			freebsd.ProcStateZombie,
			freebsd.ProcStateUnknown,
		} {
			ch <- prometheus.MustNewConstMetric(c.processByState, prometheus.GaugeValue,
				float64(stats.ByState[state]), c.host, jidStr, name, state.String())
		}
	}

	return nil
}
