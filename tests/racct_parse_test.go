//go:build freebsd && cgo

package tests

import (
	"testing"

	"github.com/zombocoder/go-freebsd-jails-exporter/internal/freebsd"
)

func TestIsRACCTEnabled(t *testing.T) {
	// Should not panic regardless of kernel config.
	enabled := freebsd.IsRACCTEnabled()
	t.Logf("RACCT enabled: %v", enabled)
}

func TestGetJailRACCT_NonexistentJail(t *testing.T) {
	if !freebsd.IsRACCTEnabled() {
		t.Skip("RACCT not enabled on this system")
	}

	_, err := freebsd.GetJailRACCT("nonexistent_jail_12345")
	if err == nil {
		t.Log("no error for nonexistent jail (may return empty results)")
	}
}
