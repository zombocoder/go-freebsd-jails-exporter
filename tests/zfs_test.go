//go:build freebsd && cgo

package tests

import (
	"testing"

	"github.com/zombocoder/go-freebsd-jails-exporter/internal/collector"
)

func TestZFSCollector(t *testing.T) {
	c := collector.NewZFSCollector(testLogger(), "testhost")

	if c.Name() != "zfs" {
		t.Fatalf("expected name 'zfs', got %q", c.Name())
	}

	count := collectAndDrain(t, c)
	t.Logf("zfs collector emitted %d metrics", count)
}
