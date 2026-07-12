---
label: unit
explanation: startup banner after AuthTest with team/bot/settings/lock
---

## Expected

- Output contains `Using config from: (none)` (CLI tokens only).
- Team name and/or team id from auth.test (`SlackTest Team` / `T024BE7LD`).
- Bot user id (`U023BECGF`) and preferably bot name (`TestSlackBot`) and/or bot_id.
- Session mode default `thread` (or session-mode marker).
- Require-mention setting visible (default true / require-mention).
- Agent-runner default `grok-tty` (or agent-runner marker).
- Lock path equals explicit `req.LockFile` (not `(none)`).
- Multi-line banner ends with trailing newline after last content line.

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
	out := resp.Stdout + resp.Stderr
	for _, want := range []string{
		"Using config from: (none)",
		slackTestTeamName,
		slackTestTeamID,
		slackTestBotUserID,
		req.LockFile,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("banner missing %q\noutput:\n%s", want, out)
		}
	}
	// Bot display name preferred; bot_id from auth is acceptable if present.
	if !strings.Contains(out, slackTestBotName) && !strings.Contains(out, slackTestAuthBotID) {
		t.Fatalf("banner missing bot name %q or bot_id %q\noutput:\n%s", slackTestBotName, slackTestAuthBotID, out)
	}
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "session") || !strings.Contains(out, "thread") {
		t.Fatalf("banner should include session-mode thread\noutput:\n%s", out)
	}
	if !strings.Contains(lower, "require-mention") && !strings.Contains(lower, "require mention") {
		t.Fatalf("banner should include require-mention setting\noutput:\n%s", out)
	}
	if !strings.Contains(out, "grok-tty") && !strings.Contains(lower, "agent-runner") {
		t.Fatalf("banner should include agent-runner (grok-tty)\noutput:\n%s", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("banner/output should end with trailing newline")
	}
}
```
