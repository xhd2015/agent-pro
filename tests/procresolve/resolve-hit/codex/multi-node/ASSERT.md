## Expected

- No error.
- `Kind` = `codex`.
- `SessionID` = fixture codex uuid from rollout path (not cmdline).
- `Confidence` = `hard`.
- `Source` = `open-files+tree`.
- `RunnerPID` = 303; `RunnerCmd` contains `codex`.
- `InputPID` = 300.
- `Tree` includes all four pids with roles:
  - 300: `input`
  - 301: `agent-run-serve`
  - 302: `other`
  - 303: `codex`

## Side Effects

- None.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoError(t, err)
	r := assertResult(t, resp)

	assertEqualInt(t, "InputPID", r.InputPID, 300)
	assertEqualString(t, "Kind", r.Kind, "codex")
	assertEqualString(t, "SessionID", r.SessionID, fixtureCodexSessionID)
	assertEqualString(t, "Confidence", r.Confidence, "hard")
	assertEqualString(t, "Source", r.Source, "open-files+tree")
	assertEqualInt(t, "RunnerPID", r.RunnerPID, 303)
	if !strings.Contains(r.RunnerCmd, "codex") {
		t.Fatalf("RunnerCmd: got %q, want substring %q", r.RunnerCmd, "codex")
	}
	if len(r.Tree) != 4 {
		t.Fatalf("Tree len: got %d, want 4; tree=%+v", len(r.Tree), r.Tree)
	}
	assertRole(t, r.Tree, 300, "input")
	assertRole(t, r.Tree, 301, "agent-run-serve")
	assertRole(t, r.Tree, 302, "other")
	assertRole(t, r.Tree, 303, "codex")
}
```
