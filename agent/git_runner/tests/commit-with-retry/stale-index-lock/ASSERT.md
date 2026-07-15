## Expected

- `CommitWithRetry` returns no error despite a stale `index.lock` at start.
- HEAD subject equals `feat: after stale lock`.
- Output does not leave the commit failed.

## Side Effects

- New commit on the temp repo branch.
- Stale `index.lock` removed before successful attempt.

## Errors

- None from `CommitWithRetry`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil Response")
	}
	if resp.CommitErr != nil {
		t.Fatalf("CommitWithRetry should succeed after stale lock removal: %v\noutput:\n%s", resp.CommitErr, resp.Output)
	}
	if resp.Subject != req.Message {
		t.Fatalf("commit subject = %q, want %q", resp.Subject, req.Message)
	}
	if strings.Contains(resp.Output, "index.lock") && strings.Contains(strings.ToLower(resp.Output), "unable to create") {
		// Successful commits may still print warnings on earlier attempts; final path must not be a hard fail.
		// With RemoveStaleIndexLock before attempt, first attempt usually succeeds with no lock noise.
		t.Logf("note: output mentioned index.lock (may be from an intermediate attempt):\n%s", resp.Output)
	}
}
```
