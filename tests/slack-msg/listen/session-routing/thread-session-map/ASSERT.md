---
label: unit
explanation: listen upserts sessions.json with session binding + config_path
---

## Expected

- `$HOME/.agent-pro/slack-local-bot/sessions.json` exists.
- Exactly one matching entry for session id `slack-channel-{channelID}`.
- Entry fields: `channel_id`, `thread_ts`, `config_path` absolute (matches listen config),
  `reply_mode` is `channel` (or equivalent channel top-level mode).
- No bot token embedded in the JSON.

## Exit Code

0

```go
import (
	"os"
	"path/filepath"
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
	sessionID := "slack-channel-" + slackTestChannelID
	doc, readErr := readSessionsJSON(t, req.HomeDir)
	if readErr != nil {
		t.Fatalf("expected sessions.json: %v", readErr)
	}
	var found *sessionMapEntry
	for i := range doc.Entries {
		if doc.Entries[i].SessionID == sessionID {
			found = &doc.Entries[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("sessions.json missing entry %q: %+v", sessionID, doc.Entries)
	}
	if found.ChannelID != slackTestChannelID {
		t.Fatalf("channel_id = %q, want %q", found.ChannelID, slackTestChannelID)
	}
	if found.ThreadTS != "1710000720.000100" {
		t.Fatalf("thread_ts = %q, want 1710000720.000100", found.ThreadTS)
	}
	if req.ConfigPath == "" {
		t.Fatal("harness expected ConfigPath set")
	}
	absCfg, _ := filepath.Abs(req.ConfigPath)
	if found.ConfigPath != req.ConfigPath && found.ConfigPath != absCfg {
		t.Fatalf("config_path = %q, want %q", found.ConfigPath, req.ConfigPath)
	}
	if found.ReplyMode != "" && found.ReplyMode != "channel" {
		t.Fatalf("reply_mode = %q, want channel (or empty default)", found.ReplyMode)
	}
	raw, _ := os.ReadFile(expectedSessionsJSONPath(req.HomeDir))
	if strings.Contains(string(raw), "xoxb-") {
		t.Fatalf("sessions.json must not embed bot token:\n%s", raw)
	}
}
```
