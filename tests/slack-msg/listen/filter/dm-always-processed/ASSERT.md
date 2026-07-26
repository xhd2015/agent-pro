---
label: unit
explanation: DM bypasses requireMention gate
---

## Expected

- Exactly one agent launch invocation.
- (Thread mode: no agent-body PostMessage required; filter only cares agent ran.)

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 1 {
		t.Fatalf("want 1 agent call, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
}
```
