# Scenario

**Feature**: `checkGrokWritable` and `OpenReady` classify grok tty-watch snapshot text

```
tty-watch snapshot <session> -> rendered scrollback text
  -> agenttty.Get("grok-tty").CheckWritable(scrollback)
  -> WritableStatus{Ready, State, Reason}
  -> (open-ready leaves) BannerDetected / OpenReady / ClassifyGrokScreen in Assert
  -> open-lifecycle readiness (orthogonal to send-queue writable)
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agenttty` registers `grok-tty` with `CheckWritable`.
- Fixtures live at `pkgs/agenttty/testdata/grok-writable/` (`grok-*.txt` + `expectations.jsonl`).
- Fixtures are **rendered** snapshot text (same bytes `tty-watch snapshot` prints), not raw PTY ANSI.
- `DOCTEST_ROOT` resolves to this tree root; module root is `DOCTEST_ROOT/../../../..`.
- Open-ready leaves require exported `OpenReady`, `ClassifyGrokScreen`, and `BannerDetected`
  (RED until implementer adds them). Shared `Run` only calls `CheckWritable` so F1 stays buildable.

## Steps

1. Root `Setup` sets default `TestdataDir` and `RepoRoot` on `Request`.
2. `Run` loads fixture bytes (single file or full table) or invokes probe export (F5).
3. Leaf `Setup` narrows `Request` (fixture name, `RunAllFixtures`, export dirs).
4. Leaf `Assert` compares writable outcomes; open-ready leaves also call exported open-ready APIs.

## Context

- Session-18 bug: full-scrollback `working` substring inside `git working tree status` false-positives
  as `busy` while prompt is idle — regression leaf must fail RED until implementer hardens detection.
- Modern SeaTalk chrome lacks legacy `Grok ›` / `GROK_TTY_BANNER`; open wait must use `OpenReady`.
- Project-directory modal false-positives legacy banner via `"grok build"`; `OpenReady` must be false.
- F5 depends on `script/debug/grok-writable-probe -export-fixtures` (implementer scope).
- F1 table asserts only `ready`/`state`/`reason` so optional JSON open-ready fields are ignored.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

const (
	fixtureFalsePositiveSession18          = "grok-after_git-idle-false-positive-session18-synth.txt"
	fixtureBusyThinking                    = "grok-regression-busy-thinking-prompt-tail.txt"
	fixtureBootEmpty                       = "grok-boot-unknown-empty-e3b0c442.txt"
	fixtureWorkspaceProjectDirectoryConfirm = "grok-workspace-project-directory-confirm.txt"
	fixtureModernStarting                  = "grok-modern-starting-session-chrome.txt"
	fixtureModernBusy                      = "grok-modern-busy-thinking-tasks.txt"
	fixtureModernIdle                      = "grok-modern-idle-input-post-turn.txt"
	fixtureLegacyAngleResponse             = "grok-regression-idle-legacy-angle-response.txt"
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

// assertOpenReadyTriplet compares precomputed open-lifecycle fields (call sites invoke
// exported agenttty.OpenReady / BannerDetected / ClassifyGrokScreen so only those leaves
// need the symbols — keeps F1 writable-only buildable).
func assertOpenReadyTriplet(t *testing.T, label string, gotLegacy, wantLegacy, gotOpen, wantOpen bool, gotClass, wantClass string) {
	t.Helper()
	if gotLegacy != wantLegacy {
		t.Fatalf("%s: banner_detected_legacy got %v want %v", label, gotLegacy, wantLegacy)
	}
	if gotOpen != wantOpen {
		t.Fatalf("%s: open_ready got %v want %v (screen_class=%q)", label, gotOpen, wantOpen, gotClass)
	}
	if wantClass != "" && gotClass != wantClass {
		t.Fatalf("%s: screen_class got %q want %q", label, gotClass, wantClass)
	}
}
```
