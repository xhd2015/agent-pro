---
label: unit
explanation: default allowFrom wildcard processes any user
---

## Expected

- One agent invocation.

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
}
```