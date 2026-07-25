## Expected

- No error.
- Hard hit unchanged: Kind=grok, SessionID=fixture grok uuid, Confidence=hard,
  Source=open-files, RunnerPID=100.
- `GrokTitle` = `fixture-grok-title`.
- `GrokModel` = `fixture-grok-model`.

## Side Effects

- None (lookup is pure inject; no disk).

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
	_ = req
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
	assertEqualString(t, "GrokTitle", r.GrokTitle, fixtureGrokTitle)
	assertEqualString(t, "GrokModel", r.GrokModel, fixtureGrokModel)
}
```
