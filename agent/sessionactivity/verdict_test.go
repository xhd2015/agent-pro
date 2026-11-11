package sessionactivity

import (
	"errors"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/dsh/consumers"
	dshsessions "github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

// verdictSources states the three observations literally, so a test says which
// one decides instead of inferring it from the filesystem.
func verdictSources(rs RunStatus, out DurableOutcome, held, writerKnown bool) dshSources {
	return dshSources{
		runStatus: rs,
		outcome:   func(string) DurableOutcome { return out },
		writers:   func(string) (bool, bool) { return held, writerKnown },
	}
}

func answers(look Lookup) RunStatus {
	return func(string) Lookup { return look }
}

// The regression: a session killed with a turn open. Its log left the turn
// open, the daemon was restarted so the registry no longer lists it, and no
// process holds the write lease. That is an interrupted error; reporting it as
// idle made every surface render a finished session.
func TestKilledOpenTurnIsInterrupted(t *testing.T) {
	id := "s-verdict-killed"
	writeDSHLog(t, id, turnStart())

	fold := DurableOutcome{Known: true, Current: true, Running: true}
	src := verdictSources(answers(Lookup{Answered: true}), fold, false, true)
	snap, err := readDSHStatus(id, src)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != ErrorKind || snap.Detail != DetailInterrupted {
		t.Fatalf("%+v, a killed turn must be an interrupted error", snap)
	}
	if snap.Outcome != consumers.OutcomeInterrupted {
		t.Fatalf("outcome=%q want %q", snap.Outcome, consumers.OutcomeInterrupted)
	}
	if snap.Unanswered || snap.Live {
		t.Fatalf("%+v, a dead turn is answered and not live", snap)
	}
	if snap.Workspace != "/workspace" {
		t.Fatalf("workspace=%q, the verdict must still carry the log's metadata", snap.Workspace)
	}
}

// The same open turn, but a process still holds the lease: the turn executes
// outside the registry and its phase must come from the log, not from the fold's
// suspicion that nobody is writing.
func TestOpenTurnWithALiveWriterKeepsItsPhase(t *testing.T) {
	id := "s-verdict-held"
	writeDSHLog(t, id, turnStart())

	fold := DurableOutcome{Known: true, Current: true, Running: true}
	src := verdictSources(answers(Lookup{Answered: true}), fold, true, true)
	snap, err := readDSHStatus(id, src)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != Thinking || !snap.Live {
		t.Fatalf("%+v, an open turn with its writer alive is running", snap)
	}
	if snap.Unanswered {
		t.Fatalf("%+v, a running turn is answered", snap)
	}
}

// A lease that cannot be observed is not a released lease: the read abstains so
// nobody deletes a live session or announces a failure it cannot prove.
func TestOpenTurnWithAnUnknownWriterAbstains(t *testing.T) {
	id := "s-verdict-writer-unknown"
	writeDSHLog(t, id, turnStart())

	fold := DurableOutcome{Known: true, Current: true, Running: true}
	src := verdictSources(answers(Lookup{Answered: true}), fold, false, false)
	snap, err := readDSHStatus(id, src)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind == Idle || snap.Kind == ErrorKind && snap.Detail == DetailInterrupted {
		t.Fatalf("%+v, an unknown writer must not decide the turn either way", snap)
	}
	if snap.Detail != DetailWriterUnknown || !snap.Unanswered {
		t.Fatalf("%+v, want the writer-unknown placeholder", snap)
	}
}

// A status answer sampled before the newest write cannot testify about the turn
// that write opened, so the lease sample taken afterwards cannot either.
func TestOpenTurnNewerThanTheAnswerAbstains(t *testing.T) {
	id := "s-verdict-log-ahead"
	writeDSHLog(t, id, turnStart())

	fold := DurableOutcome{Known: true, Current: true, Running: true, Modified: time.UnixMilli(2000)}
	look := Lookup{Answered: true, At: time.UnixMilli(1999)}
	src := verdictSources(answers(look), fold, false, true)
	snap, err := readDSHStatus(id, src)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind == Idle || snap.Kind == NotFound {
		t.Fatalf("%+v, a turn the answer predates is not decided", snap)
	}
	if snap.Detail != DetailLogAhead || !snap.Unanswered {
		t.Fatalf("%+v, want the log-ahead placeholder", snap)
	}

	// The same fold and the same lease from an answer at least as new as the log
	// do decide, which is what makes the guard above a freshness check and not a
	// blanket refusal.
	look.At = time.UnixMilli(2000)
	src = verdictSources(answers(look), fold, false, true)
	snap, err = readDSHStatus(id, src)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != ErrorKind || snap.Detail != DetailInterrupted || snap.Unanswered {
		t.Fatalf("%+v, a fresh answer plus a released lease is an interrupted turn", snap)
	}
}

// A quiescent log whose newest turn closed is answered from the fold alone: the
// tail would re-derive what the last turn/end already says, and it costs a
// multi-megabyte read for the exact sessions a dashboard polls most.
func TestSettledFoldDecidesWithoutTheTail(t *testing.T) {
	cases := []struct {
		last       string
		wantKind   Kind
		wantDetail string
	}{
		{last: consumers.OutcomeCompleted, wantKind: Idle},
		{last: "", wantKind: Idle},
		{last: consumers.OutcomeAborted, wantKind: Aborted, wantDetail: consumers.OutcomeAborted},
		{last: consumers.OutcomeInterrupted, wantKind: ErrorKind, wantDetail: consumers.OutcomeInterrupted},
		{last: consumers.OutcomeError, wantKind: ErrorKind, wantDetail: consumers.OutcomeError},
		{last: consumers.OutcomeMaxTokens, wantKind: ErrorKind, wantDetail: consumers.OutcomeMaxTokens},
	}
	for _, tc := range cases {
		t.Run("last="+tc.last, func(t *testing.T) {
			header := dshHeaderInfo{Summary: dshsessions.Summary{ID: "s1", Cwd: "/hdr", Title: "t"}, ModTime: 99}
			seams := stubDSHSeams(t, nil, errors.New("the tail must not be read"), header, nil)
			fold := DurableOutcome{Known: true, Current: true, Last: tc.last}
			src := verdictSources(answers(Lookup{Answered: true}), fold, false, true)
			snap, err := readDSHStatus("s1", src)
			if err != nil {
				t.Fatal(err)
			}
			if snap.Kind != tc.wantKind || snap.Detail != tc.wantDetail {
				t.Fatalf("%+v want %s/%q", snap, tc.wantKind, tc.wantDetail)
			}
			if snap.Outcome != tc.last || snap.Unanswered {
				t.Fatalf("%+v, a settled fold is answered and carries its outcome", snap)
			}
			if seams.tails != 0 {
				t.Fatalf("tail reads=%d, a settled fold must not read the log", seams.tails)
			}
			if seams.headers != 1 || snap.Workspace != "/hdr" {
				t.Fatalf("headers=%d workspace=%q, metadata comes from the header", seams.headers, snap.Workspace)
			}
		})
	}
}

// For a session the registry still lists, the registry wins: its answer is
// fresher than the fold and carries the live phase the grid renders.
func TestRegistryRunningWinsOverAStaleFold(t *testing.T) {
	header := dshHeaderInfo{Summary: dshsessions.Summary{ID: "s1"}}
	tail := &dshsessions.Tail{Summary: dshsessions.Summary{ID: "s1", Cwd: "/live"}}
	seams := stubDSHSeams(t, tail, nil, header, nil)
	// A fold that still says the previous turn completed must not mask a turn
	// the registry reports as running right now.
	fold := DurableOutcome{Known: true, Current: true, Last: consumers.OutcomeCompleted}
	running := Lookup{Running: true, Present: true, Answered: true}
	snap, err := readDSHStatus("s1", verdictSources(answers(running), fold, false, true))
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != Idle || !snap.Live {
		t.Fatalf("%+v, a running session is answered by its tail", snap)
	}
	if seams.tails != 1 || seams.headers != 0 {
		t.Fatalf("tails=%d headers=%d, want exactly one tail read", seams.tails, seams.headers)
	}
}

// Without a usable fold there is no evidence either way for a session the
// registry does not list. Every such answer must abstain: idle is what reported
// the killed session as finished, and error would be a failure nobody observed.
func TestUnfoldedSessionNeverReportsIdle(t *testing.T) {
	header := dshHeaderInfo{Summary: dshsessions.Summary{ID: "s1", Cwd: "/hdr"}}
	cases := []struct {
		name        string
		fold        DurableOutcome
		look        Lookup
		wantKind    Kind
		wantDetail  string
		wantTails   int
		wantHeaders int
	}{
		{
			name:        "absent and unfolded abstains",
			fold:        DurableOutcome{},
			look:        Lookup{Answered: true},
			wantKind:    ErrorKind,
			wantDetail:  DetailNotFolded,
			wantHeaders: 1,
		},
		{
			name:        "behind fold on a listed session reads the tail",
			fold:        DurableOutcome{Known: true, Current: false},
			look:        Lookup{Answered: true, Present: true},
			wantKind:    Idle,
			wantTails:   1,
			wantHeaders: 0,
		},
		{
			name:        "absent and unfolded after a partial fold abstains",
			fold:        DurableOutcome{Known: false},
			look:        Lookup{Answered: true},
			wantKind:    ErrorKind,
			wantDetail:  DetailNotFolded,
			wantHeaders: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tail := &dshsessions.Tail{Summary: dshsessions.Summary{ID: "s1", Cwd: "/live"}}
			seams := stubDSHSeams(t, tail, nil, header, nil)
			snap, err := readDSHStatus("s1", verdictSources(answers(tc.look), tc.fold, false, true))
			if err != nil {
				t.Fatal(err)
			}
			if snap.Kind != tc.wantKind {
				t.Fatalf("kind=%s want %s (%+v)", snap.Kind, tc.wantKind, snap)
			}
			if tc.wantKind == ErrorKind {
				if !snap.Unanswered || snap.Detail != tc.wantDetail {
					t.Fatalf("%+v, an undecided phase is an unanswered placeholder", snap)
				}
			}
			if seams.tails != tc.wantTails || seams.headers != tc.wantHeaders {
				t.Fatalf("tails=%d headers=%d want %d/%d", seams.tails, seams.headers, tc.wantTails, tc.wantHeaders)
			}
		})
	}
}

// not_found keeps its old meaning: the log itself is gone. Abstaining next to a
// missing log must not hide a session whose file was deleted.
func TestMissingLogStaysNotFound(t *testing.T) {
	headerErr := errors.New(`DSH session "s1" was not found under /home/.dsh/sessions`)
	seams := stubDSHSeams(t, nil, nil, dshHeaderInfo{}, headerErr)
	src := verdictSources(answers(Lookup{Answered: true}), DurableOutcome{}, false, true)
	snap, err := readDSHStatus("s1", src)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != NotFound || snap.Detail != headerErr.Error() {
		t.Fatalf("%+v want not_found with the header error", snap)
	}
	if seams.tails != 0 {
		t.Fatalf("tail reads=%d, a missing log must not be tailed", seams.tails)
	}
}

// The transcript card and the activity grid must never disagree: a session
// killed with its turn open reads as interrupted on both, and the card still
// renders the rows the log kept.
func TestFollowAgreesWithReadOnAKilledTurn(t *testing.T) {
	tail := &dshsessions.Tail{
		Summary: dshsessions.Summary{ID: "s1"},
		Rows:    []dshsessions.TranscriptRow{{Seq: 1, Kind: dshsessions.TranscriptUser, Text: "hi"}},
	}
	header := dshHeaderInfo{Summary: dshsessions.Summary{ID: "s1", Cwd: "/hdr"}}
	stubDSHSeams(t, tail, nil, header, nil)

	fold := DurableOutcome{Known: true, Current: true, Running: true}
	src := verdictSources(answers(Lookup{Answered: true}), fold, false, true)
	read, err := readDSHStatus("s1", src)
	if err != nil {
		t.Fatal(err)
	}
	snap, rows, err := followDSHStatus("s1", 0, 10, src)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != read.Kind || snap.Detail != read.Detail {
		t.Fatalf("follow=%+v read=%+v", snap, read)
	}
	if snap.Kind != ErrorKind || snap.Detail != DetailInterrupted {
		t.Fatalf("%+v, a killed turn must be interrupted on both surfaces", snap)
	}
	if len(rows) != 1 || rows[0].Text != "hi" {
		t.Fatalf("rows=%+v, the card must keep rendering a killed session", rows)
	}
}

// A session whose log vanished is the only not_found the follow path may return,
// and it must still not silently drop the card to idle.
func TestFollowMissingLogIsNotFound(t *testing.T) {
	stubDSHSeams(t, nil, errors.New("no log"), dshHeaderInfo{}, errors.New("no log"))
	src := verdictSources(answers(Lookup{Answered: true}), DurableOutcome{}, false, true)
	snap, _, err := followDSHStatus("s1", 0, 10, src)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != NotFound {
		t.Fatalf("kind=%s want %s", snap.Kind, NotFound)
	}
}

// The production probe is the composition eluc installs: a daemon that no longer
// lists the session, a fold that left the turn open, and no process holding the
// lease.
func TestDefaultProbeComposesTheInstalledHooks(t *testing.T) {
	t.Cleanup(func() {
		SetDefaultRunStatus(nil)
		SetDefaultOutcome(nil)
		SetDefaultWriter(nil)
	})
	id := "s-verdict-probe"
	writeDSHLog(t, id, turnStart())

	SetDefaultRunStatus(answers(Lookup{Answered: true}))
	SetDefaultOutcome(func(string) DurableOutcome {
		return DurableOutcome{Known: true, Current: true, Running: true}
	})
	SetDefaultWriter(func(string) (bool, bool) { return false, true })

	snap, err := Read(RunnerDSH, id)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != ErrorKind || snap.Detail != DetailInterrupted {
		t.Fatalf("%+v, the installed hooks must compose the interrupted verdict", snap)
	}
}

func TestDefaultProbeAbstainsWithoutTheOutcomeHook(t *testing.T) {
	t.Cleanup(func() { SetDefaultRunStatus(nil) })
	id := "s-verdict-probe-no-fold"
	writeDSHLog(t, id, turnStart())
	SetDefaultRunStatus(answers(Lookup{Answered: true}))

	snap, err := Read(RunnerDSH, id)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind == Idle {
		t.Fatalf("%+v, no fold means no verdict, never idle", snap)
	}
	if !snap.Unanswered || snap.Detail != DetailNotFolded {
		t.Fatalf("%+v, want the not-folded placeholder", snap)
	}
}

func TestPickOutcome(t *testing.T) {
	fold := DurableOutcome{Known: true, Current: true, Running: true, Last: "aborted", Turns: 7}
	src := DurableOutcomeSource(func(id string) DurableOutcome {
		if id != "s1" {
			t.Fatalf("asked about %q", id)
		}
		return fold
	})
	if got := pickOutcome("s1", src); got != fold {
		t.Fatalf("got %+v want %+v", got, fold)
	}
	// A nil hook is no fold at all, which the readers must read as "unknown",
	// never as a session with no open turn.
	if got := pickOutcome("s1", nil); got.Known || got.Current || got.Running || got.Last != "" {
		t.Fatalf("a nil source answered %+v", got)
	}
}

// Callers that act on a failure must be able to tell a death from a provider
// error: a liveness detail names a turn nobody answered.
func TestIsLivenessDetail(t *testing.T) {
	for _, detail := range []string{
		DetailTurnOpen,
		DetailInterrupted,
		DetailLogAhead,
		DetailWriterUnknown,
		DetailNotFolded,
		DetailUnreadable,
	} {
		if !IsLivenessDetail(detail) {
			t.Fatalf("%q is a liveness detail", detail)
		}
	}
	for _, detail := range []string{"", consumers.OutcomeCompleted, consumers.OutcomeMaxTokens, consumers.OutcomeError, consumers.OutcomeAborted, "429 rate limited"} {
		if IsLivenessDetail(detail) {
			t.Fatalf("%q is a reported failure, not a liveness detail", detail)
		}
	}
}

func TestPickWriter(t *testing.T) {
	src := WriterSource(func(id string) (bool, bool) {
		if id != "s1" {
			t.Fatalf("asked about %q", id)
		}
		return true, true
	})
	if held, known := pickWriter("s1", src); !held || !known {
		t.Fatalf("held=%v known=%v", held, known)
	}
	if held, known := pickWriter("s1", nil); held || known {
		t.Fatalf("a nil source answered held=%v known=%v", held, known)
	}
}

func TestSetDefaultDurableHooksInstallAndClear(t *testing.T) {
	t.Cleanup(func() {
		SetDefaultOutcome(nil)
		SetDefaultWriter(nil)
	})
	if defaultOutcome() != nil || defaultWriter() != nil {
		t.Fatal("a hook was installed before the test")
	}
	SetDefaultOutcome(func(string) DurableOutcome { return DurableOutcome{Known: true, Last: "error"} })
	SetDefaultWriter(func(string) (bool, bool) { return false, true })
	if got := defaultOutcome()("s1"); got.Last != "error" {
		t.Fatalf("installed outcome hook answered %+v", got)
	}
	if held, known := defaultWriter()("s1"); held || !known {
		t.Fatalf("installed writer hook answered held=%v known=%v", held, known)
	}
	SetDefaultOutcome(nil)
	SetDefaultWriter(nil)
	if defaultOutcome() != nil || defaultWriter() != nil {
		t.Fatal("a hook survived clearing")
	}
}

func TestProbeDurableHookPrecedence(t *testing.T) {
	t.Cleanup(func() {
		SetDefaultOutcome(nil)
		SetDefaultWriter(nil)
	})
	SetDefaultOutcome(func(string) DurableOutcome { return DurableOutcome{Known: true, Last: "completed"} })
	SetDefaultWriter(func(string) (bool, bool) { return false, true })

	if got := (Probe{}).outcomeSource()("s1"); got.Last != "completed" {
		t.Fatalf("probe without a field ignored the installed outcome hook: %+v", got)
	}
	if got := (Probe{Outcome: func(string) DurableOutcome { return DurableOutcome{Last: "aborted"} }}).outcomeSource()("s1"); got.Last != "aborted" {
		t.Fatalf("probe did not prefer its own outcome hook: %+v", got)
	}
	if held, known := (Probe{}).writerSource()("s1"); held || !known {
		t.Fatalf("probe without a field ignored the installed writer hook: %v/%v", held, known)
	}
	if held, known := (Probe{Writers: func(string) (bool, bool) { return true, true }}).writerSource()("s1"); !held || !known {
		t.Fatalf("probe did not prefer its own writer hook: %v/%v", held, known)
	}
}

// The hooks belong to the DSH reader only: installing them must not change how
// Grok or Codex answer, nor make a fixed reader consult them.
func TestProbeDurableHooksDoNotLeak(t *testing.T) {
	t.Cleanup(func() {
		SetDefaultOutcome(nil)
		SetDefaultWriter(nil)
	})
	SetDefaultOutcome(func(string) DurableOutcome {
		t.Fatal("outcome hook consulted for a reader that must not use it")
		return DurableOutcome{}
	})
	SetDefaultWriter(func(string) (bool, bool) {
		t.Fatal("writer hook consulted for a reader that must not use it")
		return false, false
	})
	fixed := func(id string) (Snapshot, error) { return Snapshot{SessionID: id}, nil }
	for _, runner := range []string{RunnerDSH, RunnerGrok, RunnerCodex} {
		if _, err := (Probe{DSH: fixed, Grok: fixed, Codex: fixed}).Read(runner, "s1"); err != nil {
			t.Fatal(err)
		}
	}
}
