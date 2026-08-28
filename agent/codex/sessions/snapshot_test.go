package sessions

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func writeCodexRollout(t *testing.T, home, sessionID string) {
	t.Helper()
	dir := filepath.Join(home, "sessions", "2026", "08", "01")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "rollout-2026-08-01T12-00-00-"+sessionID+".jsonl")
	body := `{"timestamp":"2026-08-01T12:00:00.000Z","type":"session_meta","payload":{"id":"` + sessionID + `","cwd":"/tmp/proj"}}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunSnapshot_CaptureExactlyOne(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff01"
	writeCodexRollout(t, home, sid)

	fake := &SnapshotFake{
		FocusFake: FocusFake{
			Procs: []FocusProc{
				{PID: 5001, PPID: 1, TTY: "/dev/ttys148", Cmd: "/usr/local/bin/codex"},
			},
			OpenFiles: map[int][]string{
				5001: {codexRolloutPath(sid)},
			},
			ITerm: []iterm2.SessionRef{
				{WindowID: "3", WindowName: "work", TabIndex: 1, SessionID: "w2t1p0", TTY: "/dev/ttys148"},
			},
		},
		ContentsByID: map[string]iterm2.ContentsResult{
			"w2t1p0": {SessionID: "w2t1p0", App: "/Applications/iTerm.app", Contents: "codex pane text"},
		},
	}
	var stdout, stderr bytes.Buffer
	if err := RunSnapshot([]string{sid}, &stdout, &stderr, home, fake.SnapshotOpts()); err != nil {
		t.Fatalf("RunSnapshot: %v stderr=%q", err, stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != "codex pane text" {
		t.Fatalf("stdout=%q", got)
	}
	if len(fake.ContentsCalls) != 1 {
		t.Fatalf("ContentsCalls=%v", fake.ContentsCalls)
	}
}

func TestRunSnapshot_NoLiveFails(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff02"
	writeCodexRollout(t, home, sid)
	fake := &SnapshotFake{
		FocusFake: FocusFake{
			Procs:     nil,
			OpenFiles: map[int][]string{},
			ITerm:     nil,
		},
	}
	err := RunSnapshot([]string{sid}, ioDiscard(), ioDiscard(), home, fake.SnapshotOpts())
	if err == nil || !strings.Contains(err.Error(), "no hosting iTerm tab") {
		t.Fatalf("err=%v", err)
	}
}

func TestRunSnapshot_Unknown(t *testing.T) {
	t.Parallel()
	err := RunSnapshot([]string{"019f283a-ffff-7fff-ffff-ffffffffff99"}, ioDiscard(), ioDiscard(), t.TempDir(), &SnapshotOpts{})
	if err == nil || !strings.Contains(err.Error(), "codex session not found") {
		t.Fatalf("err=%v", err)
	}
}

func TestRunSnapshot_TabCapture(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283b-dddd-7ddd-dddd-dddddddddddd"
	writeCodexRollout(t, home, sid)
	fake := &SnapshotFake{
		FocusFake: FocusFake{
			Procs: []FocusProc{
				{PID: 8200, PPID: 1, TTY: "/dev/ttys102", Cmd: "/usr/local/bin/codex"},
			},
			OpenFiles: map[int][]string{
				8200: {codexRolloutPath(sid)},
			},
			ITerm: []iterm2.SessionRef{
				{WindowID: "100", WindowName: "work", TabIndex: 1, SessionID: "w0t1p0:CURRENT-TAB-UUID", TTY: "/dev/ttys101"},
				{WindowID: "100", WindowName: "work", TabIndex: 2, SessionID: "w0t2p0:TAB2-UUID", TTY: "/dev/ttys102"},
			},
		},
		CurrentSessionID: "w0t1p0:CURRENT-TAB-UUID",
		ControllingTTY:   "/dev/ttys101",
		ContentsByID: map[string]iterm2.ContentsResult{
			"TAB2-UUID": {SessionID: "TAB2-UUID", App: "/Applications/iTerm.app", Contents: "codex tab pane"},
		},
	}
	var stdout bytes.Buffer
	if err := RunSnapshot([]string{"--tab", "2"}, &stdout, ioDiscard(), home, fake.SnapshotOpts()); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "codex tab pane" {
		t.Fatalf("stdout=%q", got)
	}
}

func TestRunSnapshot_AgentRunPrefer(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff03"
	writeCodexRollout(t, home, sid)
	fake := &SnapshotFake{
		AgentRunEnabled: true,
		AgentRunByID: map[string]*AgentRunSnapshotResult{
			sid: {AgentRunSessionID: "ar-1", Contents: "agent-run frame"},
		},
	}
	var stdout bytes.Buffer
	opts := fake.SnapshotOpts()
	// Force prefer path even with no Contents inject: AgentRunSnapshot is set.
	if err := RunSnapshot([]string{sid, "--json"}, &stdout, ioDiscard(), home, opts); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"source":"agent-run"`) {
		t.Fatalf("stdout=%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "agent-run frame") {
		t.Fatalf("stdout=%s", stdout.String())
	}
}

func ioDiscard() *bytes.Buffer { return &bytes.Buffer{} }
