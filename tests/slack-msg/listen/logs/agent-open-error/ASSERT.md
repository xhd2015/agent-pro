---
label: unit
explanation: agent open failure surfaces on stderr
---

## Expected

- Agent launch attempted (INVOCATION logged).
- Combined output contains a failure marker (`fail` / `error` / `failed`) related to agent open (not silent drop).

## Exit Code

0 (daemon still running until SIGTERM)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 1 {
		t.Fatalf("want 1 agent launch attempt, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
	out := strings.ToLower(resp.Stdout + resp.Stderr)
	if !(strings.Contains(out, "fail") || strings.Contains(out, "error") || strings.Contains(out, "failed")) {
		t.Fatalf("expected agent open failure log, got:\nstdout:\n%s\nstderr:\n%s", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(out, "agent") {
		t.Fatalf("failure log should mention agent:\n%s", resp.Stdout+resp.Stderr)
	}
}
```
