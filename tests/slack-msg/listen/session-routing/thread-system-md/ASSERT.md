---
label: unit
explanation: thread mode writes SYSTEM.md with send/history recipes
---

## Expected

- SYSTEM.md path:
  `$HOME/.agent-pro/slack-local-bot/sessions/slack-{channel}-{ts}/SYSTEM.md`
- File exists after first open.
- Contents include:
  - session id
  - channel id
  - thread_ts
  - `slack-msg send` recipe (reply)
  - `slack-msg history` recipe
- Contents must **not** include raw bot token (`xoxb-slacktest-token` / `xoxb-`).

## Exit Code

0

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 1 {
		t.Fatalf("want 1 agent launch, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
	sessionID := "slack-" + slackTestChannelID + "-1710000700.000100"
	path := expectedSessionSystemMDPath(req.HomeDir, sessionID)
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("expected SYSTEM.md at %s: %v\nagent: %v", path, readErr, resp.AgentInvocations)
	}
	body := string(data)
	for _, want := range []string{
		sessionID,
		slackTestChannelID,
		"1710000700.000100",
		"slack-msg send",
		"slack-msg history",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("SYSTEM.md missing %q\npath=%s\nbody:\n%s", want, path, body)
		}
	}
	if strings.Contains(body, slackTestBotToken) || strings.Contains(body, "xoxb-slacktest") {
		t.Fatalf("SYSTEM.md must not embed bot token secrets:\n%s", body)
	}
}
```
