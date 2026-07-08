# Scenario

**Feature**: `checkGrokWritable` classifies grok tty-watch snapshot text for send-queue readiness

```
tty-watch snapshot <session> -> rendered scrollback text
  -> agenttty.Get("grok-tty").CheckWritable(scrollback)
  -> WritableStatus{Ready, State, Reason}
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agenttty` registers `grok-tty` with `CheckWritable`.
- Fixtures live at `pkgs/agenttty/testdata/grok-writable/` (`grok-*.txt` + `expectations.jsonl`).
- Fixtures are **rendered** snapshot text (same bytes `tty-watch snapshot` prints), not raw PTY ANSI.
- `DOCTEST_ROOT` resolves to this tree root; module root is `DOCTEST_ROOT/../../../..`.

## Steps

1. Root `Setup` sets default `TestdataDir` and `RepoRoot` on `Request`.
2. `Run` loads fixture bytes (single file or full table) or invokes probe export (F5).
3. Leaf `Setup` narrows `Request` (fixture name, `RunAllFixtures`, export dirs).
4. Leaf `Assert` compares `WritableStatus` to expected values.

## Context

- Session-18 bug: full-scrollback `working` substring inside `git working tree status` false-positives
  as `busy` while prompt is idle — regression leaf must fail RED until implementer hardens detection.
- F5 depends on `script/debug/grok-writable-probe -export-fixtures` (implementer scope).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

const (
	fixtureFalsePositiveSession18 = "grok-after_git-idle-false-positive-session18-synth.txt"
	fixtureBusyThinking           = "grok-regression-busy-thinking-prompt-tail.txt"
	fixtureBootEmpty              = "grok-boot-unknown-empty-e3b0c442.txt"
)

func Setup(t *testing.T, req *Request) error {
	if req.RepoRoot == "" {
		req.RepoRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", "..", "..", ".."))
	}
	if req.TestdataDir == "" {
		req.TestdataDir = filepath.Join(req.RepoRoot, "pkgs", "agenttty", "testdata", "grok-writable")
	}
	return nil
}

func statusMatches(t *testing.T, label string, exp FixtureExpectation, actual agenttty.WritableStatus) {
	t.Helper()
	if actual.Ready != exp.Ready {
		t.Fatalf("%s %s: ready got %v want %v (state=%q reason=%q)", label, exp.File, actual.Ready, exp.Ready, actual.State, actual.Reason)
	}
	if actual.State != exp.State {
		t.Fatalf("%s %s: state got %q want %q (ready=%v reason=%q)", label, exp.File, actual.State, exp.State, actual.Ready, actual.Reason)
	}
	if exp.Reason != "" && actual.Reason != exp.Reason {
		t.Fatalf("%s %s: reason got %q want %q", label, exp.File, actual.Reason, exp.Reason)
	}
}

func assertWritable(t *testing.T, label string, actual agenttty.WritableStatus, wantReady bool, wantState, wantReason string) {
	t.Helper()
	if actual.Ready != wantReady {
		t.Fatalf("%s: ready got %v want %v (state=%q reason=%q)", label, actual.Ready, wantReady, actual.State, actual.Reason)
	}
	if actual.State != wantState {
		t.Fatalf("%s: state got %q want %q (ready=%v reason=%q)", label, actual.State, wantState, actual.Ready, actual.Reason)
	}
	if wantReason != "" && actual.Reason != wantReason {
		t.Fatalf("%s: reason got %q want %q", label, actual.Reason, wantReason)
	}
}
```