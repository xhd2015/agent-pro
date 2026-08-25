package knowledgesink

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestIsStaleRunning(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.Local)
	m := &Manifest{Status: statusRunning, LastPing: FormatTime(now.Add(-30 * time.Second))}
	if IsStaleRunning(m, now) {
		t.Fatal("30s should not be stale")
	}
	m.LastPing = FormatTime(now.Add(-61 * time.Second))
	if !IsStaleRunning(m, now) {
		t.Fatal("61s should be stale")
	}
	m.LastPing = ""
	if !IsStaleRunning(m, now) {
		t.Fatal("empty last_ping while running is stale")
	}
	m.Status = statusIdle
	if IsStaleRunning(m, now) {
		t.Fatal("idle must not be stale")
	}
}

func TestReconcileStaleRunning(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.Local)
	m := &Manifest{
		Version:         1,
		MarcusSessionID: "s1",
		LastSinkIndex:   -1,
		Status:          statusRunning,
		LastPing:        FormatTime(now.Add(-2 * time.Minute)),
		Pid:             1234,
	}
	if err := WriteManifest(dir, m); err != nil {
		t.Fatal(err)
	}
	ok, err := ReconcileStaleRunning(dir, m, now)
	if err != nil || !ok {
		t.Fatalf("reconcile = %v %v", ok, err)
	}
	got, err := LoadManifest(dir)
	if err != nil || got == nil {
		t.Fatal(err)
	}
	if got.Status != statusFailed || got.Error != staleError {
		t.Fatalf("got = %+v", got)
	}
}

func TestLatchAndTouchPing(t *testing.T) {
	dir := t.TempDir()
	t0 := time.Date(2026, 8, 25, 12, 0, 0, 0, time.Local)
	m, err := LatchRunning(dir, "sess", 42, t0)
	if err != nil || m == nil {
		t.Fatal(err)
	}
	if m.Status != statusRunning || m.Pid != 42 || m.LastPing != FormatTime(t0) {
		t.Fatalf("latch = %+v", m)
	}
	t1 := t0.Add(4 * time.Second)
	if err := TouchPing(dir, 42, t1); err != nil {
		t.Fatal(err)
	}
	got, _ := LoadManifest(dir)
	if got.LastPing != FormatTime(t1) {
		t.Fatalf("ping = %q", got.LastPing)
	}
}

func TestAppendLog(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.Local)
	if err := AppendLog(dir, "stdout", "hello\n", "ui", 0, now); err != nil {
		t.Fatal(err)
	}
	if err := AppendLog(dir, "stderr", "warn\n", "ui", 0, now); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(LogsPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{`"stream":"stdout"`, `"text":"hello\n"`, `"trigger":"ui"`} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %s in %s", want, s)
		}
	}
}
