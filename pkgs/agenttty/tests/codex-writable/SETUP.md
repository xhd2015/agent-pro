# Scenario

**Feature**: `checkCodexWritable` classifies codex tty-watch snapshot text for status/input readiness

```
tty-watch snapshot <session> -> rendered scrollback text
  -> agenttty.Get("codex-tty").CheckWritable(scrollback)
  -> WritableStatus{Ready, State, Reason}
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/pkgs/agenttty` registers `codex-tty` with `CheckWritable`.
- Fixtures live at `pkgs/agenttty/testdata/codex-writable/` (`codex-*.txt` + `expectations.jsonl`).
- Fixtures are **rendered** snapshot text (same bytes `tty-watch snapshot` prints), not raw PTY ANSI.
- `d.DOCTEST_ROOT` resolves to this tree root; module root is `d.DOCTEST_ROOT/../../../..`.

## Steps

1. Root `Setup` sets default `TestdataDir` and `RepoRoot` on `Request`.
2. `Run` loads fixture bytes (single file or full table) and calls `codex-tty` `CheckWritable`.
3. Leaf `Setup` narrows `Request` (fixture name or `RunAllFixtures`).
4. Leaf `Assert` compares `WritableStatus` to expected values (reason matched as substring when set).

## Context

- FetchStatus bug: **Update available** modal contains `›` on the menu option line and was
  classified `ready=true` / `state=idle`, so `/status` was injected into the modal.
- Post-fix: update modal → non-idle (`loading`); main chat `›` after boot remains idle.
- Codex **0.146.0** main-chat prompt may be `»` (U+00BB) instead of `›` (U+203A); both must
  classify as idle when present without busy/loading signals.
- `reason` in `expectations.jsonl` is a **substring** expectation (implementer-chosen full reason).

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
)

const (
	fixtureUpdateModal      = "codex-update-available-modal.txt"
	fixtureModelLoading     = "codex-update-plus-model-loading.txt"
	fixtureMainPromptMCP    = "codex-mcp-incomplete-prompt.txt"
	fixtureEmptySnapshot    = "codex-empty-snapshot.txt"
	fixtureDoubleAngleIdle  = "codex-double-angle-prompt-idle.txt"
	fixtureDoubleAngleMCP   = "codex-double-angle-mcp-incomplete.txt"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.RepoRoot == "" {
		req.RepoRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", "..", "..", ".."))
	}
	if req.TestdataDir == "" {
		req.TestdataDir = filepath.Join(req.RepoRoot, "pkgs", "agenttty", "testdata", "codex-writable")
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
	if exp.Reason != "" && !strings.Contains(actual.Reason, exp.Reason) {
		t.Fatalf("%s %s: reason got %q want substring %q", label, exp.File, actual.Reason, exp.Reason)
	}
}

func assertWritable(t *testing.T, label string, actual agenttty.WritableStatus, wantReady bool, wantState, wantReasonSubstr string) {
	t.Helper()
	if actual.Ready != wantReady {
		t.Fatalf("%s: ready got %v want %v (state=%q reason=%q)", label, actual.Ready, wantReady, actual.State, actual.Reason)
	}
	if actual.State != wantState {
		t.Fatalf("%s: state got %q want %q (ready=%v reason=%q)", label, actual.State, wantState, actual.Ready, actual.Reason)
	}
	if wantReasonSubstr != "" && !strings.Contains(actual.Reason, wantReasonSubstr) {
		t.Fatalf("%s: reason got %q want substring %q", label, actual.Reason, wantReasonSubstr)
	}
}
```
