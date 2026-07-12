---
label: unit
explanation: thread follow-up dispatches RunInteractiveOpen run (not send)
---

## Expected

- Two launch invocations.
- Both use interactive open `run` profile (`--session-id=`, `--auto-send-or-resume`,
  `--new-terminal`, `--open`).
- Neither uses bare `send` subcommand or `--keep-tty`.
- Session id is `slack-channel-{channelID}` for both.

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
	sessionID := "slack-channel-" + slackTestChannelID
	for i, line := range resp.AgentInvocations {
		for _, want := range []string{
			"run",
			"--session-id=" + sessionID,
			"--auto-send-or-resume",
			"--new-terminal",
			"--open",
		} {
			if !strings.Contains(line, want) {
				t.Fatalf("invocation %d missing %q in %q", i, want, line)
			}
		}
		if strings.Contains(line, "--keep-tty") {
			t.Fatalf("invocation %d must not use --keep-tty: %q", i, line)
		}
		if strings.Contains(line, " send ") || strings.HasPrefix(strings.TrimPrefix(line, "INVOCATION "), "send ") {
			t.Fatalf("invocation %d should not use send subcommand: %q", i, line)
		}
	}
}
```
