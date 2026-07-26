## Expected

- `GrokSyncWorkerCount()` returns **1** (not 2).
- Exactly one user `message` for `idempotent-probe-prompt`.
- `GrokSyncWorkerActive(runner, sessionID)` is true while worker runs.

## Errors

- Duplicate user events for the same pre-seeded updates line indicate overlapping workers.

## Exit Code

N/A (direct package call)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.EnsureErr != nil {
		t.Fatalf("EnsureGrokSync: %v", resp.EnsureErr)
	}
	if resp.WorkerCount != 1 {
		t.Fatalf("GrokSyncWorkerCount: got %d want 1", resp.WorkerCount)
	}
	if count := countUserMessagesByText(resp.Events, idempotentUserPrompt); count != 1 {
		t.Fatalf("user message count for %q: got %d want 1; events=%d",
			idempotentUserPrompt, count, len(resp.Events))
	}
	if !resp.WorkerActive {
		t.Fatal("expected active grok sync worker after concurrent Ensure")
	}
}
```