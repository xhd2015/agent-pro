---
label: unit
explanation: distinct ts values each open an agent
---

## Expected

- Exactly **two** agent launch invocations.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 2 {
		t.Fatalf("want 2 agent launches for different ts, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
}
```
