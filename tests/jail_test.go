//go:build freebsd && cgo

package tests

import (
	"testing"

	"github.com/zombocoder/go-freebsd-jail/jail"
	"github.com/zombocoder/go-freebsd-jails-exporter/internal/freebsd"
)

func TestJailList(t *testing.T) {
	jails, err := jail.List()
	if err != nil {
		t.Fatalf("jail.List() failed: %v", err)
	}

	t.Logf("found %d jails", len(jails))

	for _, j := range jails {
		t.Logf("jid=%d name=%s hostname=%s state=%s vnet=%v ip4=%v ip6=%v",
			j.JID, j.Name, j.Hostname, j.State, j.VNET, j.IP4, j.IP6)

		extra, err := freebsd.GetJailExtra(int(j.JID))
		if err != nil {
			t.Logf("  GetJailExtra failed: %v", err)
			continue
		}
		t.Logf("  parent=%d persist=%v osrelease=%s osreldate=%d",
			extra.Parent, extra.Persist, extra.OSRelease, extra.OSRelDate)
	}
}

func TestGetJailExtraNonexistent(t *testing.T) {
	_, err := freebsd.GetJailExtra(999999)
	if err == nil {
		t.Error("expected error for nonexistent jail JID")
	}
}
