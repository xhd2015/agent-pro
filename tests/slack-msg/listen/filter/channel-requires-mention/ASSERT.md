---
label: unit
explanation: requireMention default drops plain channel messages
---

## Expected

- No agent invocation.
- No PostMessage reply.

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
	if len(resp.AgentInvocations) != 0 {
		t.Fatalf("expected filter drop, got agent calls: %v", resp.AgentInvocations)
	}
}
```
