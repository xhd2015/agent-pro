---
label: unit
explanation: thread interactive open does not PostMessage agent body
---

## Expected

- Exactly one agent launch (open profile).
- Zero PostMessage captures (TTY owns session; SeaTalk parity).
- Agent argv includes `--open` / `--auto-send-or-resume`.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 1 {
		t.Fatalf("want 1 agent launch, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
	line := resp.AgentInvocations[0]
	if !strings.Contains(line, "--open") || !strings.Contains(line, "--auto-send-or-resume") {
		t.Fatalf("expected interactive open argv, got %q", line)
	}
	if len(resp.PostMessages) != 0 {
		t.Fatalf("thread interactive open must not PostMessage agent body, got %v", resp.PostMessages)
	}
}
```
