//go:build freebsd && cgo

package tests

import (
	"testing"

	"github.com/zombocoder/go-freebsd-jails-exporter/internal/freebsd"
)

func TestGetAllJailProcStats(t *testing.T) {
	stats, err := freebsd.GetAllJailProcStats()
	if err != nil {
		t.Fatalf("GetAllJailProcStats failed: %v", err)
	}

	t.Logf("found process stats for %d jails", len(stats))

	for jid, s := range stats {
		t.Logf("jid=%d processes=%d threads=%d", jid, s.Processes, s.Threads)
		for state, count := range s.ByState {
			t.Logf("  state=%s count=%d", state.String(), count)
		}
	}
}

func TestProcStateStrings(t *testing.T) {
	tests := []struct {
		state freebsd.ProcState
		want  string
	}{
		{freebsd.ProcStateRunning, "running"},
		{freebsd.ProcStateSleeping, "sleeping"},
		{freebsd.ProcStateStopped, "stopped"},
		{freebsd.ProcStateZombie, "zombie"},
		{freebsd.ProcStateUnknown, "unknown"},
		{freebsd.ProcState(99), "unknown"},
	}

	for _, tt := range tests {
		got := tt.state.String()
		if got != tt.want {
			t.Errorf("ProcState(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}
