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

// racctMetric maps a RACCT resource name to a Prometheus metric descriptor.
type racctMetric struct {
	resource string
	desc     *prometheus.Desc
}

type RACCTCollector struct {
	logger  *slog.Logger
	host    string
	metrics []racctMetric
}

func NewRACCTCollector(logger *slog.Logger, host string) *RACCTCollector {
	labels := []string{"host", "jid", "name"}

	makeDesc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc(name, help, labels, nil)
	}

	return &RACCTCollector{
		logger: logger,
		host:   host,
		metrics: []racctMetric{
			{"cputime", makeDesc("freebsd_jail_racct_cpu_time_seconds", "CPU time consumed in seconds.")},
			{"pcpu", makeDesc("freebsd_jail_racct_cpu_percent", "CPU usage percentage.")},
			{"memoryuse", makeDesc("freebsd_jail_racct_memory_used_bytes", "Current memory usage in bytes.")},
			{"memorylocked", makeDesc("freebsd_jail_racct_memory_locked_bytes", "Locked memory in bytes.")},
			{"vmemoryuse", makeDesc("freebsd_jail_racct_vmemory_used_bytes", "Current virtual memory usage in bytes.")},
			{"swapuse", makeDesc("freebsd_jail_racct_swap_used_bytes", "Current swap usage in bytes.")},
			{"maxproc", makeDesc("freebsd_jail_racct_processes", "Number of processes.")},
			{"nthr", makeDesc("freebsd_jail_racct_threads", "Number of threads.")},
			{"openfiles", makeDesc("freebsd_jail_racct_open_files", "Number of open files.")},
			{"datasize", makeDesc("freebsd_jail_racct_datasize_bytes", "Data segment size in bytes.")},
			{"stacksize", makeDesc("freebsd_jail_racct_stacksize_bytes", "Stack size in bytes.")},
			{"pseudoterminals", makeDesc("freebsd_jail_racct_pseudoterminals", "Number of pseudo-terminals.")},
			{"wallclock", makeDesc("freebsd_jail_racct_wallclock_seconds", "Wall clock time in seconds.")},
			{"readbps", makeDesc("freebsd_jail_racct_read_bytes_per_second", "Read throughput in bytes per second.")},
			{"writebps", makeDesc("freebsd_jail_racct_write_bytes_per_second", "Write throughput in bytes per second.")},
			{"readiops", makeDesc("freebsd_jail_racct_read_iops", "Read I/O operations per second.")},
			{"writeiops", makeDesc("freebsd_jail_racct_write_iops", "Write I/O operations per second.")},
		},
	}
}

func (c *RACCTCollector) Name() string { return "racct" }

func (c *RACCTCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m.desc
	}
}

func (c *RACCTCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !freebsd.IsRACCTEnabled() {
		c.logger.Debug("RACCT not enabled, skipping")
		return nil
	}

	jails, err := jail.List()
	if err != nil {
		return fmt.Errorf("listing jails: %w", err)
	}

	for _, j := range jails {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		jidStr := strconv.Itoa(int(j.JID))
		name := j.Name

		racct, err := freebsd.GetJailRACCT(name)
		if err != nil {
			c.logger.Warn("failed to get RACCT for jail",
				"jid", jidStr, "jail_name", name, "error", err)
			continue
		}

		for _, m := range c.metrics {
			val, ok := racct.Resources[m.resource]
			if !ok {
				continue
			}
			ch <- prometheus.MustNewConstMetric(m.desc, prometheus.GaugeValue,
				float64(val), c.host, jidStr, name)
		}
	}

	return nil
}
