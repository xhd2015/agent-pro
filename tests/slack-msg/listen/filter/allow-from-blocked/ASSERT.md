---
label: unit
explanation: allowFrom user ID gate blocks other users
---

## Expected

- No agent invocation for blocked user.

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
		t.Fatalf("expected blocked user filtered, got: %v", resp.AgentInvocations)
	}
}
```
