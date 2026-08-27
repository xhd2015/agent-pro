package sessions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreferMainOverChildSubagents_DropsChild(t *testing.T) {
	parent := "019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa"
	child := "019f283b-bbbb-7bbb-bbbb-bbbbbbbbbbbb"
	hits := []tabGrokHit{
		{SessionID: parent, RunnerPID: 1, TTY: "/dev/ttys101"},
		{SessionID: child, RunnerPID: 1, TTY: "/dev/ttys101"},
	}
	meta := func(sid string) (TabSessionMeta, bool) {
		switch sid {
		case parent:
			return TabSessionMeta{}, true
		case child:
			return TabSessionMeta{RawKind: "subagent", ParentID: parent}, true
		default:
			return TabSessionMeta{}, false
		}
	}
	got := preferMainOverChildSubagents(hits, meta)
	if len(got) != 1 || got[0].SessionID != parent {
		t.Fatalf("got %+v, want only parent %s", got, parent)
	}
}

func TestPreferMainOverChildSubagents_KeepsUnrelatedMains(t *testing.T) {
	a := "019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa"
	b := "019f283b-bbbb-7bbb-bbbb-bbbbbbbbbbbb"
	hits := []tabGrokHit{
		{SessionID: a, RunnerPID: 1, TTY: "/dev/ttys101"},
		{SessionID: b, RunnerPID: 2, TTY: "/dev/ttys101"},
	}
	meta := func(sid string) (TabSessionMeta, bool) {
		return TabSessionMeta{}, true // both main (empty kind)
	}
	got := preferMainOverChildSubagents(hits, meta)
	if len(got) != 2 {
		t.Fatalf("got %d hits, want 2 unrelated mains", len(got))
	}
}

func TestPreferMainOverChildSubagents_KeepsOrphanSubagent(t *testing.T) {
	child := "019f283b-bbbb-7bbb-bbbb-bbbbbbbbbbbb"
	hits := []tabGrokHit{{SessionID: child, RunnerPID: 1, TTY: "/dev/ttys101"}}
	meta := func(sid string) (TabSessionMeta, bool) {
		return TabSessionMeta{RawKind: "subagent", ParentID: "019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa"}, true
	}
	got := preferMainOverChildSubagents(hits, meta)
	if len(got) != 1 || got[0].SessionID != child {
		t.Fatalf("got %+v, want orphan subagent kept", got)
	}
}

func TestLookupTabSessionMeta_FromOpenPath(t *testing.T) {
	dir := t.TempDir()
	sid := "019f283b-dddd-7ddd-dddd-dddddddddddd"
	sessDir := filepath.Join(dir, sid)
	if err := os.MkdirAll(sessDir, 0o755); err != nil {
		t.Fatal(err)
	}
	summary := `{"info":{"id":"` + sid + `","cwd":"/tmp"},"session_kind":"subagent","parent_session_id":"019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa","updated_at":"2026-01-01T00:00:00Z","created_at":"2026-01-01T00:00:00Z"}`
	if err := os.WriteFile(filepath.Join(sessDir, "summary.json"), []byte(summary), 0o644); err != nil {
		t.Fatal(err)
	}
	hit := tabGrokHit{SessionID: sid, OpenPath: filepath.Join(sessDir, "events.jsonl")}
	m, ok := lookupTabSessionMeta(&TabResolveOpts{}, hit)
	if !ok {
		t.Fatal("expected meta from open path")
	}
	if m.RawKind != "subagent" || m.ParentID != "019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("meta = %+v", m)
	}
}
