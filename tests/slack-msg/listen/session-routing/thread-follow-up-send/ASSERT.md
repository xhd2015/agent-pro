---
label: unit
explanation: thread follow-up dispatches agent-run send with session id
---

## Expected

- Two agent invocations.
- First contains `run` and `--session`.
- Second contains `send` and session id `slack-{channel}-{thread_ts}`.

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
	first := resp.AgentInvocations[0]
	second := resp.AgentInvocations[1]
	if !strings.Contains(first, "run") || !strings.Contains(first, "--session") {
		t.Fatalf("first should be run --session, got %q", first)
	}
	sessionID := "slack-" + slackTestChannelID + "-1710000200.000100"
	if !strings.Contains(second, "send") {
		t.Fatalf("second should use send, got %q", second)
	}
	if !strings.Contains(second, sessionID) {
		t.Fatalf("second should reference session %q, got %q", sessionID, second)
	}
}
```
