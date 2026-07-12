---
label: unit
explanation: two channel messages with different ts share slack-channel-{channelID}
---

## Expected

- Two launch invocations (one per message; dedupe is by channel+ts, not session).
- Both use interactive open `run` profile (`--session-id=`, `--auto-send-or-resume`,
  `--new-terminal`, `--open`).
- Both use the **same** session id `slack-channel-{channelID}` (not per-ts
  `slack-{channel}-{ts}`).
- Neither uses bare `send` or `--keep-tty`.

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
	legacyPerTS1 := "slack-" + slackTestChannelID + "-1710001000.000100"
	legacyPerTS2 := "slack-" + slackTestChannelID + "-1710001000.000200"
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
		if strings.Contains(line, "--session-id="+legacyPerTS1) || strings.Contains(line, "--session-id="+legacyPerTS2) {
			t.Fatalf("invocation %d still uses legacy per-ts session id: %q", i, line)
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
