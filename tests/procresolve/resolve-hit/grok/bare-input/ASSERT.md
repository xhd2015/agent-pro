## Expected

- No error.
- `Kind` = `grok`.
- `SessionID` = `019fabcdef-1234-5678-9abc-def012345678`.
- `Confidence` = `hard`.
- `Source` contains `open-files` and does **not** require a `+tree` suffix (exactly `open-files` preferred).
- `RunnerPID` = 100 (input).
- `RunnerCmd` contains `grok`.
- `InputPID` = 100.
- `Tree` is non-empty and includes pid 100.

## Side Effects

- None (pure resolve; no process kill / no disk write).

## Errors

- None.

## Exit Code

N/A (library API).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoError(t, err)
	r := assertResult(t, resp)

	assertEqualInt(t, "InputPID", r.InputPID, 100)
	assertEqualString(t, "Kind", r.Kind, "grok")
	assertEqualString(t, "SessionID", r.SessionID, fixtureGrokSessionID)
	assertEqualString(t, "Confidence", r.Confidence, "hard")
	assertEqualString(t, "Source", r.Source, "open-files")
	assertEqualInt(t, "RunnerPID", r.RunnerPID, 100)
	if !strings.Contains(r.RunnerCmd, "grok") {
		t.Fatalf("RunnerCmd: got %q, want substring %q", r.RunnerCmd, "grok")
	}
	if len(r.Tree) == 0 {
		t.Fatal("Tree: empty, want at least the input node")
	}
	if _, ok := findNodeByPID(r.Tree, 100); !ok {
		t.Fatalf("Tree missing input pid 100: %+v", r.Tree)
	}
}
```
