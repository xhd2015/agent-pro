## Expected

- Every concurrent run exits successfully.
- Every run prints a parseable `codex-tty: session-N` line on stderr.
- Session ids are unique across concurrent processes sharing the same `AGENT_RUN_HOME`.

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
	if resp.Err != nil {
		t.Fatalf("concurrent run failed: %v\n%s", resp.Err, resp.Stdout)
	}
	if len(resp.ConcurrentSessionIDs) != req.ConcurrentRuns {
		t.Fatalf("got %d session ids, want %d: %#v", len(resp.ConcurrentSessionIDs), req.ConcurrentRuns, resp.ConcurrentSessionIDs)
	}
	seen := map[string]int{}
	for i, id := range resp.ConcurrentSessionIDs {
		if strings.TrimSpace(id) == "" {
			t.Fatalf("run %d missing codex-tty session id; stderr:\n%s", i, resp.ConcurrentStderrs[i])
		}
		if first, ok := seen[id]; ok {
			t.Fatalf("duplicate codex-tty session id %q from runs %d and %d; all ids=%v\nstderr[%d]:\n%s\nstderr[%d]:\n%s",
				id, first, i, resp.ConcurrentSessionIDs, first, resp.ConcurrentStderrs[first], i, resp.ConcurrentStderrs[i])
		}
		seen[id] = i
	}
}
```
