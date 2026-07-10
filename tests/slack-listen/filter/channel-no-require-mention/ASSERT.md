---
label: unit
explanation: --no-require-mention allows plain channel messages
---

## Expected

- One agent invocation.
- PostMessage reply.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 1 {
		t.Fatalf("want 1 agent call, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
	if len(resp.PostMessages) < 1 {
		t.Fatal("expected reply")
	}
}
```