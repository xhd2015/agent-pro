## Expected

- Exit code 0.
- Combined stdout+stderr mentions `resolve` (case-insensitive ok for word).
- Combined text mentions `--json`.
- Prefer mentioning `proc` context.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	combined := resp.Stdout + resp.Stderr
	assertContainsFold(t, combined, "resolve")
	assertContains(t, combined, "--json")
	// Soft preference: proc context
	if !strings.Contains(strings.ToLower(combined), "proc") {
		t.Fatalf("help should mention proc:\n%s", combined)
	}
}
```
