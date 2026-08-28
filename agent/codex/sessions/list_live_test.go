package sessions

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func TestRunListLive_OneHost(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff11"
	writeCodexRollout(t, home, sid)

	fake := &ListLiveFake{
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
	}
	var stdout, stderr bytes.Buffer
	opts := fake.ListLiveOpts()
	if err := RunListLive(nil, &stdout, &stderr, home, opts); err != nil {
		t.Fatalf("RunListLive: %v stderr=%q", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, sid) {
		t.Fatalf("stdout missing sid:\n%s", out)
	}
	if !strings.Contains(out, "w=3 t=1") {
		t.Fatalf("stdout missing iterm:\n%s", out)
	}
	if !strings.Contains(out, "1 sessions") {
		t.Fatalf("stdout missing footer:\n%s", out)
	}
}

func TestRunListLive_Empty(t *testing.T) {
	t.Parallel()
	fake := &ListLiveFake{
		FocusFake: FocusFake{
			Procs:     nil,
			OpenFiles: map[int][]string{},
			ITerm:     nil,
		},
	}
	var stdout bytes.Buffer
	if err := RunListLive(nil, &stdout, ioDiscard(), t.TempDir(), fake.ListLiveOpts()); err != nil {
		t.Fatalf("RunListLive: %v", err)
	}
	if !strings.Contains(stdout.String(), "0 sessions") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunListLive_OmitNoITerm(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff12"
	writeCodexRollout(t, home, sid)
	fake := &ListLiveFake{
		FocusFake: FocusFake{
			Procs: []FocusProc{
				{PID: 5001, PPID: 1, TTY: "/dev/ttys148", Cmd: "/usr/local/bin/codex"},
			},
			OpenFiles: map[int][]string{
				5001: {codexRolloutPath(sid)},
			},
			ITerm: nil, // live PID but no iTerm match
		},
	}
	var stdout bytes.Buffer
	if err := RunListLive(nil, &stdout, ioDiscard(), home, fake.ListLiveOpts()); err != nil {
		t.Fatalf("RunListLive: %v", err)
	}
	if strings.Contains(stdout.String(), sid) {
		t.Fatalf("sid should be omitted without iTerm:\n%s", stdout.String())
	}
}
