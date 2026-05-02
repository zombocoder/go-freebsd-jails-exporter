//go:build freebsd && cgo

package tests

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zombocoder/go-freebsd-jails-exporter/internal/collector"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// collectAndDrain runs a collector and drains metrics concurrently, returning the count.
func collectAndDrain(t *testing.T, c collector.Collector) int {
	t.Helper()
	ch := make(chan prometheus.Metric, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count := 0
	done := make(chan struct{})
	go func() {
		for range ch {
			count++
		}
		close(done)
	}()

	err := c.Collect(ctx, ch)
	close(ch)
	<-done

	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}
	return count
}

func TestJailsCollector(t *testing.T) {
	c := collector.NewJailsCollector(testLogger(), "testhost")
	if c.Name() != "jails" {
		t.Fatalf("expected name 'jails', got %q", c.Name())
	}
	count := collectAndDrain(t, c)
	t.Logf("jails collector emitted %d metrics", count)
}

func TestProcessesCollector(t *testing.T) {
	c := collector.NewProcessesCollector(testLogger(), "testhost")
	if c.Name() != "processes" {
		t.Fatalf("expected name 'processes', got %q", c.Name())
	}
	count := collectAndDrain(t, c)
	t.Logf("processes collector emitted %d metrics", count)
}

func TestRACCTCollector(t *testing.T) {
	c := collector.NewRACCTCollector(testLogger(), "testhost")
	if c.Name() != "racct" {
		t.Fatalf("expected name 'racct', got %q", c.Name())
	}
	count := collectAndDrain(t, c)
	t.Logf("racct collector emitted %d metrics", count)
}

func TestRegistryCollect(t *testing.T) {
	logger := testLogger()

	reg := collector.NewRegistry(logger, 10*time.Second,
		collector.NewJailsCollector(logger, "testhost"),
		collector.NewProcessesCollector(logger, "testhost"),
		collector.NewRACCTCollector(logger, "testhost"),
		collector.NewZFSCollector(logger, "testhost"),
	)

	promReg := prometheus.NewRegistry()
	promReg.MustRegister(reg)

	families, err := promReg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	for _, f := range families {
		t.Logf("metric family: %s (%d metrics)", *f.Name, len(f.Metric))
	}
}
