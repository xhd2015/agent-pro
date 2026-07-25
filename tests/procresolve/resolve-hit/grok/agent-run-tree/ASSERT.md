## Expected

- No error.
- `Kind` = `grok`.
- `SessionID` = fixture grok uuid (from Lsof path, **not** from `--session-id=ignored-cli` on agent-run).
- `Confidence` = `hard`.
- `Source` = `open-files+tree`.
- `RunnerPID` = 202, `RunnerCmd` contains `grok`.
- `InputPID` = 200.
- `Tree` has 3 nodes:
  - pid 200 Role = `input`
  - pid 201 Role = `agent-run-serve`
  - pid 202 Role = `grok`

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

	assertEqualInt(t, "InputPID", r.InputPID, 200)
	assertEqualString(t, "Kind", r.Kind, "grok")
	assertEqualString(t, "SessionID", r.SessionID, fixtureGrokSessionID)
	assertEqualString(t, "Confidence", r.Confidence, "hard")
	assertEqualString(t, "Source", r.Source, "open-files+tree")
	assertEqualInt(t, "RunnerPID", r.RunnerPID, 202)
	if !strings.Contains(r.RunnerCmd, "grok") {
		t.Fatalf("RunnerCmd: got %q, want substring %q", r.RunnerCmd, "grok")
	}
	if len(r.Tree) != 3 {
		t.Fatalf("Tree len: got %d, want 3; tree=%+v", len(r.Tree), r.Tree)
	}
	assertRole(t, r.Tree, 200, "input")
	assertRole(t, r.Tree, 201, "agent-run-serve")
	assertRole(t, r.Tree, 202, "grok")
}
```
