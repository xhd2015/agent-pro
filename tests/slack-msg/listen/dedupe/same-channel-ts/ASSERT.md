---
label: unit
explanation: dedupe app_mention+message same channel+ts to one agent launch
---

## Expected

- Exactly **one** agent launch invocation (not two).
- No requirement on PostMessage (thread mode).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 1 {
		t.Fatalf("want 1 agent launch after dual same-ts events, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
}
```
