---
label: unit
explanation: session list --limit N after sort keeps at most N data rows
---

## Expected

- Exit code 0.
- Human table with header + exactly one data row (newest session).
- Stderr empty.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	want := formatSessionListHuman([][]string{
		{sessionListNewerID, "C01ABCDEFF0", "-", "2026-07-13T07:46:00Z", "hello from slack"},
	})
	if resp.Stdout != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, resp.Stdout)
	}
	if strings.Contains(resp.Stdout, sessionListOlderID) {
		t.Fatalf("limited list must not include older session %q:\n%s", sessionListOlderID, resp.Stdout)
	}
}
```
