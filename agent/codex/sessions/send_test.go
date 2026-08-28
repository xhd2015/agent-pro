package sessions

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func TestRunSend_ITermHost(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff31"
	writeCodexRollout(t, home, sid)

	fake := &SendFake{
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
	opts := fake.SendOpts()
	opts.NoAgentRun = true
	var stdout, stderr bytes.Buffer
	if err := RunSend([]string{"hello", "--session-id", sid, "--no-agent-run"}, &stdout, &stderr, home, opts); err != nil {
		t.Fatalf("RunSend: %v stderr=%q", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "sent to session "+sid) {
		t.Fatalf("stdout=%q", stdout.String())
	}
	// Codex submit: type without newline, then two bare Enter writes after settle.
	if len(fake.SendCalls) != 3 {
		t.Fatalf("SendCalls=%v", fake.SendCalls)
	}
	if !strings.Contains(fake.SendCalls[0].Text, "hello") {
		t.Fatalf("sent text=%q", fake.SendCalls[0].Text)
	}
	if !fake.SendCalls[0].Opts.NoSubmit {
		t.Fatalf("first write must type without newline: %+v", fake.SendCalls[0].Opts)
	}
	for i, c := range fake.SendCalls[1:] {
		if c.Text != "" || c.Opts.NoSubmit {
			t.Fatalf("enter[%d]=%+v", i, c)
		}
	}
	if fake.SleepCalls < 2 {
		t.Fatalf("SleepCalls=%d want >=2 settles", fake.SleepCalls)
	}
}

func TestRunSend_NoSubmitStagesOnly(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff33"
	writeCodexRollout(t, home, sid)

	fake := &SendFake{
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
	opts := fake.SendOpts()
	opts.NoAgentRun = true
	if err := RunSend([]string{"draft", "--session-id", sid, "--no-agent-run", "--no-submit"}, ioDiscard(), ioDiscard(), home, opts); err != nil {
		t.Fatal(err)
	}
	if len(fake.SendCalls) != 1 {
		t.Fatalf("SendCalls=%v", fake.SendCalls)
	}
	if !fake.SendCalls[0].Opts.NoSubmit {
		t.Fatalf("want staged: %+v", fake.SendCalls[0].Opts)
	}
}

func TestRunSend_MissingPayload(t *testing.T) {
	t.Parallel()
	err := RunSend([]string{"--session-id", "019f283a-ffff-7fff-ffff-ffffffffff31"}, ioDiscard(), ioDiscard(), t.TempDir(), &SendOpts{})
	if err == nil || !strings.Contains(err.Error(), "missing text or key") {
		t.Fatalf("err=%v", err)
	}
}

func TestRunSend_NoHost(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff32"
	writeCodexRollout(t, home, sid)
	fake := &SendFake{
		FocusFake: FocusFake{
			Procs:     nil,
			OpenFiles: map[int][]string{},
			ITerm:     nil,
		},
	}
	opts := fake.SendOpts()
	opts.NoAgentRun = true
	err := RunSend([]string{"hi", "--session-id", sid, "--no-agent-run"}, ioDiscard(), ioDiscard(), home, opts)
	if err == nil || !strings.Contains(err.Error(), "no hosting iTerm tab") {
		t.Fatalf("err=%v", err)
	}
}
