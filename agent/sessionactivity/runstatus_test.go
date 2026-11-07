package sessionactivity

import "testing"

func TestPickLive(t *testing.T) {
	unknown := RunStatus(func(string) (bool, bool) { return false, false })
	running := RunStatus(func(string) (bool, bool) { return true, true })
	idle := RunStatus(func(string) (bool, bool) { return false, true })
	cases := []struct {
		name                  string
		leaseLive, leaseKnown bool
		rs                    RunStatus
		wantLive, wantKnown   bool
	}{
		{name: "no hook keeps a live lease", leaseLive: true, leaseKnown: true, wantLive: true, wantKnown: true},
		{name: "no hook keeps an idle lease", leaseLive: false, leaseKnown: true, wantLive: false, wantKnown: true},
		{name: "no hook keeps an unknown lease", wantLive: false, wantKnown: false},
		{name: "running agent is live even without a lease", rs: running, wantLive: true, wantKnown: true},
		{name: "idle agent overrides a held lease", leaseLive: true, leaseKnown: true, rs: idle, wantLive: false, wantKnown: true},
		{name: "unknown id falls back to a held lease", leaseLive: true, leaseKnown: true, rs: unknown, wantLive: true, wantKnown: true},
		{name: "unknown id and unknown lease stay unknown", rs: unknown, wantLive: false, wantKnown: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			live, known := pickLive("s1", tc.leaseLive, tc.leaseKnown, tc.rs)
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

// dshLiveness must prefer an answered hook and otherwise report whatever the
// process lease cache knows.
func TestDSHLivenessFallsBackToLease(t *testing.T) {
	running := RunStatus(func(string) (bool, bool) { return true, true })
	if live, known := dshLiveness("no-such-session", running); !live || !known {
		t.Fatalf("answered hook: live=%v known=%v", live, known)
	}
	ids, leaseKnown := dshLiveIDs()
	wantLive := leaseKnown && ids["no-such-session"]
	if live, known := dshLiveness("no-such-session", nil); live != wantLive || known != leaseKnown {
		t.Fatalf("no hook: live=%v known=%v, want live=%v known=%v", live, known, wantLive, leaseKnown)
	}
}
