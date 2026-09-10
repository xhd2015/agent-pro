package idle

import (
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/tty/detection/occupied"
)

func TestWatchdog_logsTickResetSoftExitShutdown(t *testing.T) {
	var events []Event
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := t0
	snap := "stable"
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		SessionID: "sess-log",
		Timeout:   10 * time.Millisecond,
		Grace:     5 * time.Millisecond,
		Now:       func() time.Time { return now },
		Snapshot:  func() (string, error) { return snap, nil },
		ProbeOccupied: func() occupied.Status { return occupied.Empty },
		Log: func(e Event) {
			events = append(events, e)
		},
		SoftExit: func() {},
		Shutdown: func() {},
	})
	w.LogArmed()

	w.Tick() // hit1
	now = t0.Add(5 * time.Millisecond)
	snap = "changed"
	w.Tick() // reset changed
	if got := lastEvent(events, EventReset); got == nil || got.Reason != ResetChanged || got.SnapTail == "" {
		t.Fatalf("want reset=changed with snap_tail; events=%v", eventNames(events))
	}
	snap = "changed"
	now = t0.Add(10 * time.Millisecond)
	w.Tick() // hit1 after reset
	now = t0.Add(15 * time.Millisecond)
	w.Tick() // hit2
	now = t0.Add(25 * time.Millisecond)
	w.Tick() // hit3 + soft_exit
	if got := lastEvent(events, EventSoftExit); got == nil || got.Hits < 3 || got.SnapTail == "" {
		t.Fatalf("want soft_exit with tail; events=%v", eventNames(events))
	}
	now = t0.Add(35 * time.Millisecond)
	w.Tick() // shutdown via grace
	if got := lastEvent(events, EventShutdown); got == nil {
		t.Fatalf("want shutdown; events=%v", eventNames(events))
	}
	if lastEvent(events, EventArmed) == nil || lastEvent(events, EventTick) == nil {
		t.Fatalf("missing armed/tick; events=%v", eventNames(events))
	}
}

func TestWatchdog_logsNotReadyReset(t *testing.T) {
	var events []Event
	w := New(true, Policy{ExitOnIdle: true}, Watchdog{
		Snapshot:      func() (string, error) { return "busy", nil },
		ProbeOccupied: func() occupied.Status { return occupied.Empty },
		ReadyStatus: func(string) (bool, string) {
			return false, "busy:still working"
		},
		Log: func(e Event) { events = append(events, e) },
	})
	w.Tick()
	got := lastEvent(events, EventReset)
	if got == nil || got.Reason != ResetNotReady {
		t.Fatalf("want not_ready reset; events=%v", eventNames(events))
	}
	if got.ReadyState != "busy:still working" || got.Ready == nil || *got.Ready {
		t.Fatalf("ready fields: %+v", got)
	}
}

func lastEvent(events []Event, name string) *Event {
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Event == name {
			e := events[i]
			return &e
		}
	}
	return nil
}

func eventNames(events []Event) []string {
	out := make([]string, len(events))
	for i, e := range events {
		out[i] = e.Event
		if e.Reason != "" {
			out[i] += ":" + e.Reason
		}
	}
	return out
}
