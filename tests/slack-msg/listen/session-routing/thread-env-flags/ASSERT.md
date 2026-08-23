---
label: unit
explanation: listen agent argv carries -e SLACK_MSG_SESSION_ID and -e SLACK_MSG_CONFIG
---

## Expected

- One agent launch (interactive open).
- Argv includes `-e` with `SLACK_MSG_SESSION_ID=<sessionID>` (adjacent or combined).
- Argv includes `-e` with `SLACK_MSG_CONFIG=<abs config>` matching listen config.
- Still uses `--session-id=` open profile.
- Must not pass token secrets as env values.

## Exit Code

0

```go
import (
	"os"
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
	wantSID := "SLACK_MSG_SESSION_ID=" + sessionID
	wantCfg := "SLACK_MSG_CONFIG=" + req.ConfigPath
	// Prefer inline -e; long env may spill to --env-file=<path> (no -e KEY=VALUE).
	envBlob := line
	if i := strings.Index(line, "--env-file="); i >= 0 {
		rest := line[i+len("--env-file="):]
		path := rest
		if j := strings.IndexAny(rest, " \t"); j >= 0 {
			path = rest[:j]
		}
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read --env-file %q: %v (line=%q)", path, err, line)
		}
		envBlob = string(b)
	} else if !strings.Contains(line, "-e") && !strings.Contains(line, "--env") {
		t.Fatalf("agent argv must use -e/--env or --env-file for session env: %q", line)
	}
	if !strings.Contains(envBlob, wantSID) {
		t.Fatalf("agent env missing %q in argv/env-file; line=%q env=%q", wantSID, line, envBlob)
	}
	if !strings.Contains(envBlob, wantCfg) {
		t.Fatalf("agent env missing %q in argv/env-file; line=%q env=%q", wantCfg, line, envBlob)
	}
	if !strings.Contains(line, "--session-id="+sessionID) {
		t.Fatalf("agent argv missing --session-id=%s in %q", sessionID, line)
	}
	if strings.Contains(line, "xoxb-") || strings.Contains(line, slackTestBotToken) {
		t.Fatalf("agent argv must not embed bot token: %q", line)
	}
}
```
