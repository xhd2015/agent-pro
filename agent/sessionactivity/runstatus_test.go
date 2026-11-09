package sessionactivity

import (
	"errors"
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

func TestKindFromRunWithoutLiveness(t *testing.T) {
	tail := &dshsessions.Tail{}
	cases := []struct {
		name string
		look Lookup
		want Kind
	}{
		{name: "unanswered is error", look: Lookup{}, want: ErrorKind},
		{name: "absent is not_found", look: Lookup{Answered: true}, want: NotFound},
		{name: "false classifies an empty tail as idle", look: Lookup{Answered: true, Present: true}, want: Idle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := snapshotFromTail(tail, tc.look).Kind; got != tc.want {
				t.Fatalf("kind=%s want %s", got, tc.want)
			}
		})
	}
}

// dshSeams stubs the two DSH paths so a test can prove which one ran. A session
// the answered map does not list may stop at the header, so the 45 ids a
// daemon never loaded cost no tail; a listed session always owns its phase.
type dshSeams struct {
	tails   int
	headers int
}

func stubDSHSeams(t *testing.T, tail *dshsessions.Tail, tailErr error, hdr dshHeaderInfo, hdrErr error) *dshSeams {
	t.Helper()
	seams := &dshSeams{}
	prevTail, prevHeader := dshTailFn, dshHeaderFn
	dshTailFn = func(string) (*dshsessions.Tail, error) {
		seams.tails++
		return tail, tailErr
	}
	dshHeaderFn = func(string) (dshHeaderInfo, error) {
		seams.headers++
		return hdr, hdrErr
	}
	t.Cleanup(func() { dshTailFn, dshHeaderFn = prevTail, prevHeader })
	return seams
}

func TestReadDSHStatusTailsOnlyLoadedSessions(t *testing.T) {
	emptyTail := &dshsessions.Tail{Summary: dshsessions.Summary{ID: "s1", Cwd: "/live", Title: "live"}}
	header := dshHeaderInfo{
		Summary: dshsessions.Summary{ID: "s1", Cwd: "/hdr", Title: "hdr"},
		ModTime: 1234,
	}
	running := RunStatus(func(string) Lookup { return Lookup{Running: true, Present: true, Answered: true} })
	loadedIdle := RunStatus(func(string) Lookup { return Lookup{Present: true, Answered: true} })
	absent := RunStatus(func(string) Lookup { return Lookup{Answered: true} })
	cases := []struct {
		name        string
		rs          RunStatus
		wantKind    Kind
		wantLive    bool
		wantTails   int
		wantHeaders int
		wantDetail  string
		// wantUnanswered marks a placeholder snapshot: the RPC did not answer,
		// so the kind is a fallback rather than the session's real phase.
		wantUnanswered bool
	}{
		{
			name: "running reads the log", rs: running,
			wantKind: Idle, wantLive: true, wantTails: 1,
		},
		{
			// Listed, so liveness is known and the tail still owns the last word
			// on how the last turn ended: an abort or a failed turn must survive
			// the process exiting.
			name: "loaded but not running classifies the log", rs: loadedIdle,
			wantKind: Idle, wantTails: 1,
		},
		{
			// The regression this rule fixes: the daemon answering a map that
			// omits the id means "not loaded", not "no log".
			name: "absent from an answered map is idle without the log", rs: absent,
			wantKind: Idle, wantHeaders: 1,
		},
		{
			name: "unanswered rpc is error without touching disk", rs: nil,
			wantKind: ErrorKind, wantDetail: "dsh status unreachable", wantUnanswered: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seams := stubDSHSeams(t, emptyTail, nil, header, nil)
			snap, err := readDSHStatus("s1", tc.rs)
			if err != nil {
				t.Fatal(err)
			}
			if snap.Kind != tc.wantKind {
				t.Fatalf("kind=%s want %s", snap.Kind, tc.wantKind)
			}
			if snap.Live != tc.wantLive {
				t.Fatalf("live=%v want %v", snap.Live, tc.wantLive)
			}
			if seams.tails != tc.wantTails {
				t.Fatalf("tail reads=%d want %d", seams.tails, tc.wantTails)
			}
			if seams.headers != tc.wantHeaders {
				t.Fatalf("header reads=%d want %d", seams.headers, tc.wantHeaders)
			}
			if tc.wantDetail != "" && snap.Detail != tc.wantDetail {
				t.Fatalf("detail=%q want %q", snap.Detail, tc.wantDetail)
			}
			if snap.Unanswered != tc.wantUnanswered {
				t.Fatalf("unanswered=%v want %v", snap.Unanswered, tc.wantUnanswered)
			}
		})
	}
}

func TestReadDSHStatusMetadataFollowsThePath(t *testing.T) {
	emptyTail := &dshsessions.Tail{Summary: dshsessions.Summary{ID: "s1", Cwd: "/live", Title: "live"}}
	header := dshHeaderInfo{
		Summary: dshsessions.Summary{ID: "s1", Cwd: "/hdr", Title: "hdr"},
		ModTime: 1234,
	}
	running := RunStatus(func(string) Lookup { return Lookup{Running: true, Present: true, Answered: true} })
	absent := RunStatus(func(string) Lookup { return Lookup{Answered: true} })

	stubDSHSeams(t, emptyTail, nil, header, nil)
	live, err := readDSHStatus("s1", running)
	if err != nil {
		t.Fatal(err)
	}
	if live.Workspace != "/live" {
		t.Fatalf("running session took its workspace from %q, not the log", live.Workspace)
	}

	stopped, err := readDSHStatus("s1", absent)
	if err != nil {
		t.Fatal(err)
	}
	if stopped.Workspace != "/hdr" || stopped.Title != "hdr" {
		t.Fatalf("stopped session metadata %+v did not come from the header", stopped)
	}
	if stopped.UpdatedAt != 1234 {
		t.Fatalf("updatedAt=%d, want the log mtime 1234", stopped.UpdatedAt)
	}
}

func TestReadDSHStatusNotFoundRequiresAMissingLog(t *testing.T) {
	headerErr := errors.New(`DSH session "s1" was not found under /home/.dsh/sessions`)
	seams := stubDSHSeams(t, nil, nil, dshHeaderInfo{}, headerErr)
	absent := RunStatus(func(string) Lookup { return Lookup{Answered: true} })

	snap, err := readDSHStatus("s1", absent)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != NotFound {
		t.Fatalf("kind=%s want %s", snap.Kind, NotFound)
	}
	if snap.Detail != headerErr.Error() {
		t.Fatalf("detail=%q want the header error", snap.Detail)
	}
	if seams.tails != 0 {
		t.Fatalf("tail reads=%d, a missing log must not be tailed", seams.tails)
	}
}

func TestFollowDSHStatusKeepsRowsForStoppedSessions(t *testing.T) {
	tail := &dshsessions.Tail{
		Summary: dshsessions.Summary{ID: "s1"},
		Rows:    []dshsessions.TranscriptRow{{Seq: 1, Kind: dshsessions.TranscriptUser, Text: "hi"}},
	}
	header := dshHeaderInfo{Summary: dshsessions.Summary{ID: "s1", Cwd: "/hdr"}}
	absent := RunStatus(func(string) Lookup { return Lookup{Answered: true} })

	seams := stubDSHSeams(t, tail, nil, header, nil)
	snap, rows, err := followDSHStatus("s1", 0, 10, absent)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != Idle || snap.Live {
		t.Fatalf("stopped follow reported %s live=%v", snap.Kind, snap.Live)
	}
	if len(rows) != 1 || rows[0].Text != "hi" {
		t.Fatalf("rows=%+v, the transcript card must keep rendering a stopped session", rows)
	}
	if seams.tails != 1 {
		t.Fatalf("tail reads=%d want 1", seams.tails)
	}

	// A log that disappeared is the only not_found, and it must still not
	// silently drop to idle.
	stubDSHSeams(t, nil, errors.New("no log"), dshHeaderInfo{}, errors.New("no log"))
	snap, _, err = followDSHStatus("s1", 0, 10, absent)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != NotFound {
		t.Fatalf("kind=%s want %s", snap.Kind, NotFound)
	}
}

// A listed session keeps the tail's verdict even after it stops, so the
// transcript card and the activity grid agree on a failed or aborted turn.
func TestFollowDSHStatusClassifiesALoadedStoppedSession(t *testing.T) {
	tail := &dshsessions.Tail{
		Summary: dshsessions.Summary{ID: "s1", Cwd: "/live"},
		Rows:    []dshsessions.TranscriptRow{{Seq: 1, Kind: dshsessions.TranscriptUser, Text: "hi"}},
	}
	header := dshHeaderInfo{Summary: dshsessions.Summary{ID: "s1", Cwd: "/hdr"}}
	loadedIdle := RunStatus(func(string) Lookup { return Lookup{Present: true, Answered: true} })

	stubDSHSeams(t, tail, nil, header, nil)
	snap, rows, err := followDSHStatus("s1", 0, 10, loadedIdle)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Workspace != "/live" {
		t.Fatalf("workspace=%q, a listed session answers from its tail", snap.Workspace)
	}
	if snap.Kind != Idle || snap.Live {
		t.Fatalf("stopped follow reported %s live=%v", snap.Kind, snap.Live)
	}
	if len(rows) != 1 {
		t.Fatalf("rows=%+v want 1", rows)
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
