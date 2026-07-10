---
label: unit
explanation: --channel filter drops events from non-listed channels
---

## Expected

- No agent invocation for excluded channel.

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 0 {
		t.Fatalf("expected channel filter drop, got: %v", resp.AgentInvocations)
	}
}
```