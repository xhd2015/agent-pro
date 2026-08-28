package sessions

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

const (
	fixtureCodexSID    = "019f283b-cccc-7ccc-cccc-cccccccccccc"
	fixtureTabCodexSID = "019f283b-dddd-7ddd-dddd-dddddddddddd"
)

func codexRolloutPath(sessionID string) string {
	return "/Users/fixture/.codex/sessions/2026/08/01/rollout-2026-08-01T12-00-00-" + sessionID + ".jsonl"
}

func TestRunResolve_TabHit(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	opts := &ResolveOpts{
		Stdout: &stdout,
		Stderr: &stderr,
		PID:    6000,
		ListProcs: func() []procresolve.Proc {
			return nil
		},
		Lsof: func(int) []string { return nil },
		ListFocusProcs: func() []FocusProc {
			return []FocusProc{
				{PID: 8100, PPID: 1, TTY: "/dev/ttys101", Cmd: "/bin/bash"},
				{PID: 8200, PPID: 1, TTY: "/dev/ttys102", Cmd: "/usr/local/bin/codex"},
				{PID: 9100, PPID: 1, TTY: "/dev/ttys103", Cmd: "/bin/bash"},
			}
		},
		ListITerm: func() ([]iterm2.SessionRef, error) {
			return []iterm2.SessionRef{
				{WindowID: "100", WindowName: "work", TabIndex: 1, SessionID: "w0t1p0:CURRENT-TAB-UUID", TTY: "/dev/ttys101", Name: "current"},
				{WindowID: "100", WindowName: "work", TabIndex: 2, SessionID: "w0t2p0:TAB2-UUID", TTY: "/dev/ttys102", Name: "codex-tab"},
				{WindowID: "100", WindowName: "work", TabIndex: 3, SessionID: "w0t3p0:TAB3-UUID", TTY: "/dev/ttys103", Name: "bash-only"},
			}, nil
		},
		CurrentSessionID: func() string { return "w0t1p0:CURRENT-TAB-UUID" },
		ControllingTTY:   func() string { return "/dev/ttys101" },
		AncestorTTYs:     func() []string { return nil },
	}
	// Override Lsof used by tab path (same field as ancestor).
	opts.Lsof = func(pid int) []string {
		if pid == 8200 {
			return []string{codexRolloutPath(fixtureTabCodexSID)}
		}
		return nil
	}

	if err := RunResolve([]string{"--tab", "2"}, opts); err != nil {
		t.Fatalf("RunResolve: %v stderr=%q", err, stderr.String())
	}
	got := strings.TrimSpace(stdout.String())
	if got != fixtureTabCodexSID {
		t.Fatalf("stdout=%q want %q", got, fixtureTabCodexSID)
	}
}

func TestRunResolve_TabMissGrokOnly(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	opts := &ResolveOpts{
		Stdout: &stdout,
		Stderr: &stderr,
		ListFocusProcs: func() []FocusProc {
			return []FocusProc{
				{PID: 8200, PPID: 1, TTY: "/dev/ttys102", Cmd: "/usr/local/bin/grok"},
			}
		},
		ListITerm: func() ([]iterm2.SessionRef, error) {
			return []iterm2.SessionRef{
				{WindowID: "100", WindowName: "work", TabIndex: 1, SessionID: "w0t1p0:CURRENT-TAB-UUID", TTY: "/dev/ttys101"},
				{WindowID: "100", WindowName: "work", TabIndex: 2, SessionID: "w0t2p0:TAB2-UUID", TTY: "/dev/ttys102"},
			}, nil
		},
		CurrentSessionID: func() string { return "w0t1p0:CURRENT-TAB-UUID" },
		ControllingTTY:   func() string { return "/dev/ttys101" },
		Lsof: func(pid int) []string {
			if pid == 8200 {
				return []string{"/Users/fixture/.grok/sessions/%2Ftmp%2Fproj/019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa/events.jsonl"}
			}
			return nil
		},
	}
	err := RunResolve([]string{"--tab", "2"}, opts)
	if err == nil {
		t.Fatalf("want error, got stdout=%q", stdout.String())
	}
	if !strings.Contains(err.Error(), "no codex session on tab") {
		t.Fatalf("err=%v", err)
	}
}

func TestRunResolve_AncestorHit(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	procs := []procresolve.Proc{
		{PID: 100, PPID: 1, Cmd: "/usr/local/bin/codex"},
		{PID: 200, PPID: 100, Cmd: "/bin/bash"},
		{PID: 300, PPID: 200, Cmd: "/usr/local/bin/agent-pro"},
	}
	opts := &ResolveOpts{
		Stdout: &stdout,
		Stderr: &stderr,
		PID:    300,
		ListProcs: func() []procresolve.Proc {
			return procs
		},
		Lsof: func(pid int) []string {
			if pid == 100 {
				return []string{codexRolloutPath(fixtureCodexSID)}
			}
			return nil
		},
	}
	if err := RunResolve(nil, opts); err != nil {
		t.Fatalf("RunResolve: %v", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != fixtureCodexSID {
		t.Fatalf("stdout=%q want %q", got, fixtureCodexSID)
	}
}

func TestRunResolve_NoAncestor(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	opts := &ResolveOpts{
		Stdout: &stdout,
		PID:    300,
		ListProcs: func() []procresolve.Proc {
			return []procresolve.Proc{
				{PID: 200, PPID: 1, Cmd: "/bin/bash"},
				{PID: 300, PPID: 200, Cmd: "/usr/local/bin/agent-pro"},
			}
		},
		Lsof: func(int) []string { return nil },
	}
	err := RunResolve(nil, opts)
	if err == nil || !strings.Contains(err.Error(), "no ancestor codex") {
		t.Fatalf("err=%v", err)
	}
}

func TestRunResolve_PidAndTab(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	err := RunResolve([]string{"--pid", "1", "--tab", "1"}, &ResolveOpts{Stdout: &stdout})
	if err == nil || !strings.Contains(err.Error(), "--pid and --tab/--tab-index cannot be specified together") {
		t.Fatalf("err=%v", err)
	}
}
