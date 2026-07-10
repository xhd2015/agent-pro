---
label: unit
explanation: filtered bot-self message should not invoke mock agent
---

## Expected

- No `INVOCATION` lines in agent log.
- No PostMessage reply.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 0 {
		t.Fatalf("expected no agent calls, got %v", resp.AgentInvocations)
	}
	if len(resp.PostMessages) != 0 {
		t.Fatalf("expected no replies, got %v", resp.PostMessages)
	}
}
```