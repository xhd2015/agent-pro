## Expected

- `GrokSyncWorkerActive` is true before reconcile attempt.
- `EventLineCountAfter == EventLineCountBefore` (no duplicate sync).
- Exactly one user message for `reconcile skip worker prompt`.

## Errors

- Event count increasing after reconcile while worker active indicates missing skip guard.

## Exit Code

N/A (direct package call)

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.WorkerActive {
		t.Fatal("precondition: worker must be active before reconcile")
	}
	if resp.EventLineCountAfter != resp.EventLineCountBefore {
		t.Fatalf("reconcile must not double-sync: before=%d after=%d events=%d",
			resp.EventLineCountBefore, resp.EventLineCountAfter, len(resp.Events))
	}
	if count := countUserMessagesByText(resp.Events, reconcileSkipPrompt); count != 1 {
		t.Fatalf("user prompt %q: want 1 got %d (duplicate sync?)", reconcileSkipPrompt, count)
	}
}
```