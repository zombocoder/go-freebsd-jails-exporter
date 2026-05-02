//go:build freebsd && cgo

package collector

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zombocoder/go-freebsd-jail/jail"
	"github.com/zombocoder/go-freebsd-libzfs/zfs"
)

type ZFSCollector struct {
	logger *slog.Logger
	host   string

	used          *prometheus.Desc
	available     *prometheus.Desc
	referenced    *prometheus.Desc
	logicalUsed   *prometheus.Desc
	quota         *prometheus.Desc
	refquota      *prometheus.Desc
	reservation   *prometheus.Desc
	refreserv     *prometheus.Desc
	compressratio *prometheus.Desc
	usedBySnap    *prometheus.Desc
	usedByDS      *prometheus.Desc
	usedByChild   *prometheus.Desc
}

func NewZFSCollector(logger *slog.Logger, host string) *ZFSCollector {
	labels := []string{"host", "jid", "name", "dataset"}

	makeDesc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc(name, help, labels, nil)
	}

	return &ZFSCollector{
		logger:        logger,
		host:          host,
		used:          makeDesc("freebsd_jail_zfs_used_bytes", "Space used by the dataset in bytes."),
		available:     makeDesc("freebsd_jail_zfs_available_bytes", "Space available to the dataset in bytes."),
		referenced:    makeDesc("freebsd_jail_zfs_referenced_bytes", "Space referenced by the dataset in bytes."),
		logicalUsed:   makeDesc("freebsd_jail_zfs_logical_used_bytes", "Logical space used by the dataset in bytes."),
		quota:         makeDesc("freebsd_jail_zfs_quota_bytes", "Quota for the dataset in bytes (0 = none)."),
		refquota:      makeDesc("freebsd_jail_zfs_refquota_bytes", "Reference quota for the dataset in bytes (0 = none)."),
		reservation:   makeDesc("freebsd_jail_zfs_reservation_bytes", "Reservation for the dataset in bytes (0 = none)."),
		refreserv:     makeDesc("freebsd_jail_zfs_refreservation_bytes", "Reference reservation for the dataset in bytes (0 = none)."),
		compressratio: makeDesc("freebsd_jail_zfs_compressratio", "Compression ratio achieved (1.0 = no compression)."),
		usedBySnap:    makeDesc("freebsd_jail_zfs_used_by_snapshots_bytes", "Space used by snapshots of the dataset."),
		usedByDS:      makeDesc("freebsd_jail_zfs_used_by_dataset_bytes", "Space used by the dataset itself."),
		usedByChild:   makeDesc("freebsd_jail_zfs_used_by_children_bytes", "Space used by children of the dataset."),
	}
}

func (c *ZFSCollector) Name() string { return "zfs" }

func (c *ZFSCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.used
	ch <- c.available
	ch <- c.referenced
	ch <- c.logicalUsed
	ch <- c.quota
	ch <- c.refquota
	ch <- c.reservation
	ch <- c.refreserv
	ch <- c.compressratio
	ch <- c.usedBySnap
	ch <- c.usedByDS
	ch <- c.usedByChild
}

func (c *ZFSCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	jails, err := jail.List()
	if err != nil {
		return fmt.Errorf("listing jails: %w", err)
	}

	zfsClient, err := zfs.New(ctx)
	if err != nil {
		return fmt.Errorf("creating zfs client: %w", err)
	}
	defer zfsClient.Close()

	// Build mountpoint -> dataset name lookup.
	datasets, err := zfsClient.List(ctx, true)
	if err != nil {
		return fmt.Errorf("listing zfs datasets: %w", err)
	}

	mountToDataset := make(map[string]string)
	for _, ds := range datasets {
		if !ds.IsFilesystem() {
			continue
		}
		mp, err := zfsClient.GetStringProperty(ctx, ds.Name, zfs.PropertyNameMountpoint)
		if err != nil || mp == "" || mp == "none" || mp == "-" {
			continue
		}
		mountToDataset[mp] = ds.Name
	}

	// Track datasets we've already emitted to avoid duplicates when multiple
	// jails share the same path/dataset.
	emitted := make(map[string]bool)

	for _, j := range jails {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		jidStr := strconv.Itoa(int(j.JID))
		jailName := j.Name
		dsName, ok := mountToDataset[j.Path]
		if !ok {
			c.logger.Debug("no ZFS dataset found for jail path",
				"jail_name", jailName, "path", j.Path)
			continue
		}

		// For shared datasets, only emit once per dataset but label with each jail.
		// If jails share a path, they all get the same ZFS metrics.
		cacheKey := dsName + ":" + jidStr
		if emitted[cacheKey] {
			continue
		}
		emitted[cacheKey] = true

		c.collectDataset(ctx, zfsClient, ch, jidStr, jailName, dsName)
	}

	return nil
}

func (c *ZFSCollector) collectDataset(ctx context.Context, client *zfs.Client, ch chan<- prometheus.Metric, jid, jailName, dsName string) {
	labels := []string{c.host, jid, jailName, dsName}

	emitSize := func(desc *prometheus.Desc, propName string) {
		s, err := client.GetStringProperty(ctx, dsName, propName)
		if err != nil || s == "" || s == "-" || s == "none" {
			return
		}
		v, err := parseZFSSize(s)
		if err != nil {
			return
		}
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v, labels...)
	}

	emitSize(c.used, zfs.PropertyNameUsed)
	emitSize(c.available, zfs.PropertyNameAvail)
	emitSize(c.referenced, "refer")
	emitSize(c.usedBySnap, zfs.PropertyNameUsedsnap)
	emitSize(c.usedByDS, zfs.PropertyNameUsedds)
	emitSize(c.usedByChild, zfs.PropertyNameUsedchild)
	emitSize(c.logicalUsed, "logicalused")

	// Quota/reservation: "none" means not set, emit as 0.
	emitQuota := func(desc *prometheus.Desc, propName string) {
		s, err := client.GetStringProperty(ctx, dsName, propName)
		if err != nil || s == "" || s == "-" {
			return
		}
		if s == "none" || s == "0" {
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 0, labels...)
			return
		}
		v, err := parseZFSSize(s)
		if err != nil {
			return
		}
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v, labels...)
	}

	emitQuota(c.quota, zfs.PropertyNameQuota)
	emitQuota(c.refquota, zfs.PropertyNameRefquota)
	emitQuota(c.reservation, zfs.PropertyNameReservation)
	emitQuota(c.refreserv, zfs.PropertyNameRefreserv)

	// Compressratio is returned as a string like "1.86x".
	ratioStr, err := client.GetStringProperty(ctx, dsName, zfs.PropertyNameCompressratio)
	if err == nil && ratioStr != "" {
		s := ratioStr
		if len(s) > 0 && s[len(s)-1] == 'x' {
			s = s[:len(s)-1]
		}
		if ratio, err := strconv.ParseFloat(s, 64); err == nil {
			ch <- prometheus.MustNewConstMetric(c.compressratio, prometheus.GaugeValue, ratio, labels...)
		}
	}
}

// parseZFSSize converts human-readable ZFS sizes (e.g. "3.94M", "202G", "1024") to bytes.
func parseZFSSize(s string) (float64, error) {
	if s == "" || s == "-" || s == "none" {
		return 0, fmt.Errorf("empty value")
	}

	// Try plain number first.
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v, nil
	}

	if len(s) < 2 {
		return 0, fmt.Errorf("invalid size: %s", s)
	}

	suffix := s[len(s)-1]
	numStr := s[:len(s)-1]
	v, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	switch suffix {
	case 'B':
		return v, nil
	case 'K':
		return v * 1024, nil
	case 'M':
		return v * 1024 * 1024, nil
	case 'G':
		return v * 1024 * 1024 * 1024, nil
	case 'T':
		return v * 1024 * 1024 * 1024 * 1024, nil
	case 'P':
		return v * 1024 * 1024 * 1024 * 1024 * 1024, nil
	case 'E':
		return v * 1024 * 1024 * 1024 * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unknown suffix: %c", suffix)
	}
}
