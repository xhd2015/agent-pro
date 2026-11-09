package sessionactivity

import (
	"testing"

	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

func TestPickLookup(t *testing.T) {
	unknown := RunStatus(func(string) Lookup { return Lookup{} })
	running := RunStatus(func(string) Lookup { return Lookup{Running: true, Present: true, Answered: true} })
	idle := RunStatus(func(string) Lookup { return Lookup{Present: true, Answered: true} })
	absent := RunStatus(func(string) Lookup { return Lookup{Answered: true} })
	cases := []struct {
		name string
		rs   RunStatus
		want Lookup
	}{
		{name: "no hook is unanswered", want: Lookup{}},
		{name: "running agent", rs: running, want: Lookup{Running: true, Present: true, Answered: true}},
		{name: "idle agent", rs: idle, want: Lookup{Present: true, Answered: true}},
		{name: "unanswered rpc", rs: unknown, want: Lookup{}},
		{name: "id missing from a successful map", rs: absent, want: Lookup{Answered: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := pickLookup("s1", tc.rs)
			if got != tc.want {
				t.Fatalf("got %+v want %+v", got, tc.want)
			}
		})
	}
}

func TestKindFromRunIgnoresTailWhenNotRunning(t *testing.T) {
	tail := &dshsessions.Tail{}
	cases := []struct {
		name string
		look Lookup
		want Kind
	}{
		{name: "unanswered is error", look: Lookup{}, want: ErrorKind},
		{name: "absent is not_found", look: Lookup{Answered: true}, want: NotFound},
		{name: "false is idle", look: Lookup{Answered: true, Present: true}, want: Idle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := snapshotFromTail(tail, tc.look)
			if got.Kind != tc.want {
				t.Fatalf("kind=%s want %s", got.Kind, tc.want)
			}
			if got.Live {
				t.Fatal("finished/absent/error must not be live")
			}
		})
	}
}

func TestSetDefaultRunStatusInstallsAndClears(t *testing.T) {
	t.Cleanup(func() { SetDefaultRunStatus(nil) })

	if defaultRunStatus() != nil {
		t.Fatal("a hook was installed before the test")
	}
	want := RunStatus(func(string) Lookup { return Lookup{Running: true, Present: true, Answered: true} })
	SetDefaultRunStatus(want)
	if got := defaultRunStatus(); got == nil {
		t.Fatal("installed hook is missing")
	} else if look := got("s1"); !look.Running || !look.Present || !look.Answered {
		t.Fatalf("installed hook answered %+v", look)
	}
	SetDefaultRunStatus(nil)
	if defaultRunStatus() != nil {
		t.Fatal("hook survived clearing")
	}
}

func TestProbeRunStatusHookPrecedence(t *testing.T) {
	t.Cleanup(func() { SetDefaultRunStatus(nil) })
	installed := RunStatus(func(string) Lookup { return Lookup{Present: true, Answered: true} })
	own := RunStatus(func(string) Lookup { return Lookup{Running: true, Present: true, Answered: true} })

	SetDefaultRunStatus(installed)
	if got := (Probe{}).runStatusHook(); got == nil {
		t.Fatal("probe without a field ignored the installed hook")
	} else if look := got("s1"); look.Running {
		t.Fatal("probe without a field did not use the installed hook")
	}
	if got := (Probe{RunStatus: own}).runStatusHook(); got == nil {
		t.Fatal("probe field was ignored")
	} else if look := got("s1"); !look.Running {
		t.Fatal("probe did not prefer its own hook")
	}
}

func TestProbeRunStatusDoesNotLeak(t *testing.T) {
	panicHook := RunStatus(func(string) Lookup {
		t.Fatal("hook consulted for a reader that must not use it")
		return Lookup{}
	})
	fixed := func(id string) (Snapshot, error) { return Snapshot{SessionID: id, Live: true}, nil }
	if _, err := (Probe{DSH: fixed, RunStatus: panicHook}).Read("dsh", "s1"); err != nil {
		t.Fatal(err)
	}
	if _, err := (Probe{Grok: fixed, RunStatus: panicHook}).Read("grok", "s1"); err != nil {
		t.Fatal(err)
	}
	if _, err := (Probe{Codex: fixed, RunStatus: panicHook}).Read("codex", "s1"); err != nil {
		t.Fatal(err)
	}
}
