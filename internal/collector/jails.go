//go:build freebsd && cgo

package collector

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zombocoder/go-freebsd-jail/jail"
	"github.com/zombocoder/go-freebsd-jails-exporter/internal/freebsd"
)

type JailsCollector struct {
	logger *slog.Logger
	host   string

	jailInfo      *prometheus.Desc
	jailState     *prometheus.Desc
	jailParentJID *prometheus.Desc
	jailPersist   *prometheus.Desc
	jailVnet      *prometheus.Desc
	jailIP4Addrs  *prometheus.Desc
	jailIP6Addrs  *prometheus.Desc
}

func NewJailsCollector(logger *slog.Logger, host string) *JailsCollector {
	labels := []string{"host", "jid", "name"}

	return &JailsCollector{
		logger: logger,
		host:   host,
		jailInfo: prometheus.NewDesc(
			"freebsd_jail_info",
			"Information about a FreeBSD jail.",
			[]string{"host", "jid", "name", "hostname", "path", "osrelease", "osreldate"},
			nil,
		),
		jailState: prometheus.NewDesc(
			"freebsd_jail_state",
			"Current state of a FreeBSD jail.",
			[]string{"host", "jid", "name", "state"},
			nil,
		),
		jailParentJID: prometheus.NewDesc(
			"freebsd_jail_parent_jid",
			"Parent JID of a FreeBSD jail.",
			labels, nil,
		),
		jailPersist: prometheus.NewDesc(
			"freebsd_jail_persist",
			"Whether the jail has the persist flag set.",
			labels, nil,
		),
		jailVnet: prometheus.NewDesc(
			"freebsd_jail_vnet",
			"Whether the jail has VNET enabled.",
			labels, nil,
		),
		jailIP4Addrs: prometheus.NewDesc(
			"freebsd_jail_ip4_addresses",
			"Number of IPv4 addresses assigned to the jail.",
			labels, nil,
		),
		jailIP6Addrs: prometheus.NewDesc(
			"freebsd_jail_ip6_addresses",
			"Number of IPv6 addresses assigned to the jail.",
			labels, nil,
		),
	}
}

func (c *JailsCollector) Name() string { return "jails" }

func (c *JailsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.jailInfo
	ch <- c.jailState
	ch <- c.jailParentJID
	ch <- c.jailPersist
	ch <- c.jailVnet
	ch <- c.jailIP4Addrs
	ch <- c.jailIP6Addrs
}

func (c *JailsCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
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

		jid := strconv.Itoa(int(j.JID))
		name := j.Name

		// Get extra params (parent, persist, osrelease, osreldate).
		extra, extraErr := freebsd.GetJailExtra(int(j.JID))
		if extraErr != nil {
			c.logger.Warn("failed to get extra jail params",
				"jid", jid, "jail_name", name, "error", extraErr)
			extra = &freebsd.JailExtra{}
		}

		// Persist from library's Params map as fallback.
		persist := extra.Persist
		if !persist {
			if p, ok := j.Params["persist"]; ok && p == "1" {
				persist = true
			}
		}

		osrelease := extra.OSRelease
		osreldate := strconv.Itoa(extra.OSRelDate)

		ch <- prometheus.MustNewConstMetric(c.jailInfo, prometheus.GaugeValue, 1,
			c.host, jid, name, j.Hostname, j.Path, osrelease, osreldate)

		state := string(j.State)
		ch <- prometheus.MustNewConstMetric(c.jailState, prometheus.GaugeValue, 1,
			c.host, jid, name, state)

		ch <- prometheus.MustNewConstMetric(c.jailParentJID, prometheus.GaugeValue,
			float64(extra.Parent), c.host, jid, name)

		persistVal := 0.0
		if persist {
			persistVal = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.jailPersist, prometheus.GaugeValue,
			persistVal, c.host, jid, name)

		vnetVal := 0.0
		if j.VNET {
			vnetVal = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.jailVnet, prometheus.GaugeValue,
			vnetVal, c.host, jid, name)

		ch <- prometheus.MustNewConstMetric(c.jailIP4Addrs, prometheus.GaugeValue,
			float64(countAddrs(j.IP4)), c.host, jid, name)
		ch <- prometheus.MustNewConstMetric(c.jailIP6Addrs, prometheus.GaugeValue,
			float64(countAddrs(j.IP6)), c.host, jid, name)
	}

	return nil
}

func countAddrs(addrs []string) int {
	count := 0
	for _, a := range addrs {
		if strings.TrimSpace(a) != "" {
			count++
		}
	}
	return count
}
