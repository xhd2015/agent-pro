---
label: unit
explanation: session update --json full entry with session_id, agent_session_id, dir
---

## Expected

- Exit code 0.
- JSON object is the full updated map entry.
- Includes both `session_id` and `agent_session_id` (equal).
- `dir` is absolute workspace path.
- Preserved channel_id / config_path / preview; updated_at bumped.
- Trailing newline; stderr empty.
- sessions.json matches stdout dir.

## Exit Code

0

```go
import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	var doc struct {
		SessionID          string `json:"session_id"`
		AgentSessionID     string `json:"agent_session_id"`
		ChannelID          string `json:"channel_id"`
		ThreadTS           string `json:"thread_ts"`
		ConfigPath         string `json:"config_path"`
		Dir                string `json:"dir"`
		Kind               string `json:"kind"`
		ReplyMode          string `json:"reply_mode"`
		CreatedAt          string `json:"created_at"`
		UpdatedAt          string `json:"updated_at"`
		LastMessagePreview string `json:"last_message_preview"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &doc); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, resp.Stdout)
	}
	abs, absErr := filepath.Abs(filepath.Join(req.WorkDir, "agent-workspace"))
	if absErr != nil {
		t.Fatalf("abs: %v", absErr)
	}
	if doc.SessionID != sessionUpdateFixtureID {
		t.Fatalf("session_id = %q, want %q", doc.SessionID, sessionUpdateFixtureID)
	}
	if doc.AgentSessionID != doc.SessionID {
		t.Fatalf("agent_session_id = %q, want equal to session_id", doc.AgentSessionID)
	}
	if doc.Dir != abs {
		t.Fatalf("dir = %q, want %q", doc.Dir, abs)
	}
	if doc.ChannelID != slackTestChannelID {
		t.Fatalf("channel_id = %q", doc.ChannelID)
	}
	if doc.ConfigPath != "/tmp/slack-update-cfg.json" {
		t.Fatalf("config_path = %q", doc.ConfigPath)
	}
	if doc.LastMessagePreview != "before update" {
		t.Fatalf("last_message_preview = %q", doc.LastMessagePreview)
	}
	if doc.UpdatedAt == "" || doc.UpdatedAt == "2026-07-10T10:00:00Z" {
		t.Fatalf("updated_at should be bumped, got %q", doc.UpdatedAt)
	}
	if resp.Stdout == "" || resp.Stdout[len(resp.Stdout)-1] != '\n' {
		t.Fatalf("stdout must end with trailing newline, got %q", resp.Stdout)
	}
	stored, readErr := readSessionsJSON(t, req.HomeDir)
	if readErr != nil {
		t.Fatalf("read sessions.json: %v", readErr)
	}
	found := false
	for _, e := range stored.Entries {
		if e.SessionID == sessionUpdateFixtureID {
			found = true
			if e.Dir != abs {
				t.Fatalf("persisted dir = %q, want %q", e.Dir, abs)
			}
		}
	}
	if !found {
		t.Fatalf("entry missing after update")
	}
}
```
