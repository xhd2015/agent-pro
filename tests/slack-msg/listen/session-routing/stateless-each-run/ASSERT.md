---
label: unit
explanation: stateless mode never uses send subcommand
---

## Expected

- Two agent invocations.
- Both use `run`; neither uses `send`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 2 {
		t.Fatalf("want 2 invocations, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
	for i, line := range resp.AgentInvocations {
		if !strings.Contains(line, "run") {
			t.Fatalf("invocation %d should use run, got %q", i, line)
		}
		if strings.Contains(line, " send ") {
			t.Fatalf("stateless should not use send, got %q", line)
		}
	}
}
```
