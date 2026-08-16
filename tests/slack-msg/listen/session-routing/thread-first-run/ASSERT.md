---
label: unit
explanation: thread mode first message dispatches RunInteractiveOpen open profile
---

## Expected

- Exactly one launch invocation (tty status polls not counted).
- Argv contains `run`, `--session-id=slack-channel-{channelID}`, `--auto-send-or-resume`,
  `--new-terminal`, `--open`.
- Argv does **not** contain `--keep-tty` or bare `send` subcommand.

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
		t.Fatalf("want 1 invocation, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
	line := resp.AgentInvocations[0]
	sessionID := "slack-channel-" + slackTestChannelID
	for _, want := range []string{
		"run",
		"--session-id=" + sessionID,
		"--auto-send-or-resume",
		"--open",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("agent argv missing %q in %q", want, line)
		}
	}
	if strings.Contains(line, "--keep-tty") {
		t.Fatalf("interactive open must not use --keep-tty: %q", line)
	}
	// Bare send subcommand (legacy follow-up path), not --auto-send-or-resume.
	if strings.Contains(line, " send ") || strings.HasPrefix(strings.TrimPrefix(line, "INVOCATION "), "send ") {
		t.Fatalf("first thread message should not use send subcommand: %q", line)
	}
}
```
