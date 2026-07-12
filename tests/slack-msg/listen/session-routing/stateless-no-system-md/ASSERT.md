---
label: unit
explanation: stateless mode must not create SYSTEM.md session playbook
---

## Expected

- One agent launch (stateless run).
- SYSTEM.md for the would-be session id does **not** exist under HomeDir.
- Prompt need not include open-inject SYSTEM.md path (stateless keeps capture PostMessage behavior).

## Exit Code

0

```go
import (
	"os"
	"path/filepath"
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
	sessionID := "slack-channel-" + slackTestChannelID
	path := expectedSessionSystemMDPath(req.HomeDir, sessionID)
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("stateless must not write SYSTEM.md at %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat SYSTEM.md: %v", err)
	}
	// No sessions tree required either.
	sessionsRoot := filepath.Join(req.HomeDir, filepath.FromSlash(defaultSlackLocalBotRelDir), "sessions")
	if entries, err := os.ReadDir(sessionsRoot); err == nil && len(entries) > 0 {
		t.Fatalf("stateless must not create session dirs under %s: %v", sessionsRoot, entries)
	}
	line := resp.AgentInvocations[0]
	if strings.Contains(line, "SYSTEM.md") {
		t.Fatalf("stateless prompt should not reference SYSTEM.md: %q", line)
	}
}
```
