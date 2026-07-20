package view

import (
	"fmt"
	"net"
	"testing"
)

func TestListenLocalPreferredDefaultAndSkipBusy(t *testing.T) {
	// Bind an ephemeral to discover a free base, then occupy base and expect +1.
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("probe: %v", err)
	}
	base := probe.Addr().(*net.TCPAddr).Port
	_ = probe.Close()

	// Occupy base.
	busy, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", base))
	if err != nil {
		t.Fatalf("occupy: %v", err)
	}
	defer busy.Close()

	ln, err := listenLocalPreferred(base)
	if err != nil {
		t.Fatalf("listenLocalPreferred: %v", err)
	}
	defer ln.Close()
	got := ln.Addr().(*net.TCPAddr).Port
	if got == base {
		t.Fatalf("got busy port %d", got)
	}
	if got < base || got >= base+ViewPortAttempts {
		t.Fatalf("port %d outside range %d..%d", got, base, base+ViewPortAttempts-1)
	}
}

func TestListenLocalPreferredZeroUsesDefault(t *testing.T) {
	// May fail if 61781-61880 all busy; try and skip if so.
	ln, err := listenLocalPreferred(0)
	if err != nil {
		t.Skipf("default port range busy: %v", err)
	}
	defer ln.Close()
	got := ln.Addr().(*net.TCPAddr).Port
	if got < DefaultViewPort || got >= DefaultViewPort+ViewPortAttempts {
		t.Fatalf("port %d outside default range", got)
	}
}
