package sessionactivity

import "testing"

func TestPickLive(t *testing.T) {
	unknown := RunStatus(func(string) (bool, bool) { return false, false })
	running := RunStatus(func(string) (bool, bool) { return true, true })
	idle := RunStatus(func(string) (bool, bool) { return false, true })
	cases := []struct {
		name                string
		rs                  RunStatus
		wantLive, wantKnown bool
	}{
		{name: "no hook is unknown, not a lock probe", wantLive: false, wantKnown: false},
		{name: "running agent is live", rs: running, wantLive: true, wantKnown: true},
		{name: "idle agent is not live", rs: idle, wantLive: false, wantKnown: true},
		{name: "unknown id stays unknown", rs: unknown, wantLive: false, wantKnown: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			live, known := pickLive("s1", tc.rs)
			if live != tc.wantLive || known != tc.wantKnown {
				t.Fatalf("got live=%v known=%v, want live=%v known=%v", live, known, tc.wantLive, tc.wantKnown)
			}
		})
	}
}

func TestSetDefaultRunStatusInstallsAndClears(t *testing.T) {
	t.Cleanup(func() { SetDefaultRunStatus(nil) })

	if defaultRunStatus() != nil {
		t.Fatal("a hook was installed before the test")
	}
	want := RunStatus(func(string) (bool, bool) { return true, true })
	SetDefaultRunStatus(want)
	if got := defaultRunStatus(); got == nil {
		t.Fatal("installed hook is missing")
	} else if live, known := got("s1"); !live || !known {
		t.Fatalf("installed hook answered live=%v known=%v", live, known)
	}
	SetDefaultRunStatus(nil)
	if defaultRunStatus() != nil {
		t.Fatal("hook survived clearing")
	}
}

func TestProbeRunStatusHookPrecedence(t *testing.T) {
	t.Cleanup(func() { SetDefaultRunStatus(nil) })
	installed := RunStatus(func(string) (bool, bool) { return false, true })
	own := RunStatus(func(string) (bool, bool) { return true, true })

	SetDefaultRunStatus(installed)
	if got := (Probe{}).runStatusHook(); got == nil {
		t.Fatal("probe without a field ignored the installed hook")
	} else if live, _ := got("s1"); live {
		t.Fatal("probe without a field did not use the installed hook")
	}
	if got := (Probe{RunStatus: own}).runStatusHook(); got == nil {
		t.Fatal("probe field was ignored")
	} else if live, _ := got("s1"); !live {
		t.Fatal("probe did not prefer its own hook")
	}
}

// The hook answers only the default DSH reader: an explicit DSH reader owns
// liveness itself, and Grok/Codex never consult it.
func TestProbeRunStatusDoesNotLeak(t *testing.T) {
	panicHook := RunStatus(func(string) (bool, bool) {
		t.Fatal("hook consulted for a reader that must not use it")
		return false, false
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

func TestDSHLivenessUsesOnlyRunStatus(t *testing.T) {
	running := RunStatus(func(string) (bool, bool) { return true, true })
	if live, known := dshLiveness("no-such-session", running); !live || !known {
		t.Fatalf("answered hook: live=%v known=%v", live, known)
	}
	if live, known := dshLiveness("no-such-session", nil); live || known {
		t.Fatalf("no hook must not probe session.lock: live=%v known=%v", live, known)
	}
	unknown := RunStatus(func(string) (bool, bool) { return false, false })
	if live, known := dshLiveness("no-such-session", unknown); live || known {
		t.Fatalf("silent hook must not probe session.lock: live=%v known=%v", live, known)
	}
}
