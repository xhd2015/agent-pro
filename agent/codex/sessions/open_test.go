package sessions

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func TestRunOpen_FocusExactlyOne(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff21"
	writeCodexRollout(t, home, sid)

	fake := &OpenFake{
		FocusFake: FocusFake{
			Procs: []FocusProc{
				{PID: 5001, PPID: 1, TTY: "/dev/ttys148", Cmd: "/usr/local/bin/codex"},
			},
			OpenFiles: map[int][]string{
				5001: {codexRolloutPath(sid)},
			},
			ITerm: []iterm2.SessionRef{
				{WindowID: "3", WindowName: "work", TabIndex: 2, SessionID: "w2t2p0", TTY: "/dev/ttys148"},
			},
		},
	}
	var stdout, stderr bytes.Buffer
	if err := RunOpen([]string{sid}, &stdout, &stderr, home, fake.OpenOpts()); err != nil {
		t.Fatalf("RunOpen: %v stderr=%q", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "focused: window 3, tab 2") {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if len(fake.Focused) != 1 || fake.Focused[0] != "w2t2p0" {
		t.Fatalf("Focused=%v", fake.Focused)
	}
}

func TestRunOpen_DryRunResume(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff22"
	proj := filepath.Join(home, "proj")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}
	// Rollout with cwd = proj
	dir := filepath.Join(home, "sessions", "2026", "08", "01")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "rollout-2026-08-01T12-00-00-"+sid+".jsonl")
	body := `{"timestamp":"2026-08-01T12:00:00.000Z","type":"session_meta","payload":{"id":"` + sid + `","cwd":"` + proj + `"}}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	fake := &OpenFake{
		FocusFake: FocusFake{
			Procs:     nil,
			OpenFiles: map[int][]string{},
			ITerm:     nil,
		},
	}
	opts := fake.OpenOpts()
	opts.NoAgentRun = true
	var stdout bytes.Buffer
	if err := RunOpen([]string{sid, "--dry-run", "--no-agent-run"}, &stdout, ioDiscard(), home, opts); err != nil {
		t.Fatalf("RunOpen: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "Would open new iTerm2 window") {
		t.Fatalf("stdout=%q", out)
	}
	if !strings.Contains(out, "codex resume "+sid) && !strings.Contains(out, "resume "+sid) {
		t.Fatalf("want resume argv in stdout:\n%s", out)
	}
	if len(fake.Opened) != 0 {
		t.Fatalf("dry-run must not open: %v", fake.Opened)
	}
}

func TestRunOpen_Unknown(t *testing.T) {
	t.Parallel()
	err := RunOpen([]string{"019f283a-ffff-7fff-ffff-ffffffffff99"}, ioDiscard(), ioDiscard(), t.TempDir(), &OpenOpts{})
	if err == nil || !strings.Contains(err.Error(), "codex session not found") {
		t.Fatalf("err=%v", err)
	}
}
