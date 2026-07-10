---
label: unit
explanation: thread mode first message dispatches agent-run run --keep-tty --session
---

## Expected

- Agent log contains `run`, `--keep-tty`, `--session`, and `slack-` + channel + thread ts.

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
	if len(resp.AgentInvocations) != 1 {
		t.Fatalf("want 1 invocation, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
	line := resp.AgentInvocations[0]
	for _, want := range []string{"run", "--keep-tty", "--session", "slack-" + slackTestChannelID} {
		if !strings.Contains(line, want) {
			t.Fatalf("agent argv missing %q in %q", want, line)
		}
	}
	if strings.Contains(line, " send ") {
		t.Fatalf("first thread message should not use send: %q", line)
	}
}
```
