## Expected

- No error.
- Hard hit still works: Kind=grok, SessionID=fixture, Confidence=hard.
- `GrokTitle` is empty string.
- `GrokModel` is empty string.
- Must **not** equal the inject fixture title/model (enrich gate off).

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
	assertEqualString(t, "GrokTitle", r.GrokTitle, "")
	assertEqualString(t, "GrokModel", r.GrokModel, "")
}
```
