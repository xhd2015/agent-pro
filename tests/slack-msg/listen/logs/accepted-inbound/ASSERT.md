---
label: unit
explanation: operator log for accepted inbound + agent open
---

## Expected

- Combined output (prefer stderr) includes:
  - event kind (`app_mention`)
  - user display name (`spengler`) or user id if display unavailable
  - channel id
  - message ts
  - user text excerpt (`log me please`)
- Agent open/start marker (e.g. contains `agent` and `open` / `launch` / `start`).

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
	// Prefer stderr for operator logs; allow stdout fallback.
	out := resp.Stderr
	if !strings.Contains(out, slackTestChannelID) {
		out = resp.Stdout + resp.Stderr
	}
	for _, want := range []string{
		"app_mention",
		slackTestChannelID,
		"1710000600.000100",
		"log me please",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("operator log missing %q\nstderr:\n%s\nstdout:\n%s", want, resp.Stderr, resp.Stdout)
		}
	}
	if !strings.Contains(out, slackTestUserDisplayName) && !strings.Contains(out, slackTestUserID) {
		t.Fatalf("operator log missing user display %q or id %q\n%s", slackTestUserDisplayName, slackTestUserID, out)
	}
	lower := strings.ToLower(out)
	if !(strings.Contains(lower, "agent") && (strings.Contains(lower, "open") || strings.Contains(lower, "launch") || strings.Contains(lower, "start") || strings.Contains(lower, "run"))) {
		t.Fatalf("expected agent open/start log line\n%s", out)
	}
}
```
