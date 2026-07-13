---
label: unit
explanation: session update --dir stores abs path, OK line, preserves other map fields
---

## Expected

- Exit code 0.
- Stdout: `OK session=<id> dir=<abs>\n`.
- Map entry: `dir` is absolute path equal to requested workspace; channel/config/preview preserved.
- `updated_at` changed (newer than seed `2026-07-10T10:00:00Z`).
- Stderr empty.

## Exit Code

0

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	wantDir := filepath.Join(req.WorkDir, "agent-workspace")
	abs, absErr := filepath.Abs(wantDir)
	if absErr != nil {
		t.Fatalf("abs: %v", absErr)
	}
	wantOut := fmt.Sprintf("OK session=%s dir=%s\n", sessionUpdateFixtureID, abs)
	if resp.Stdout != wantOut {
		t.Fatalf("stdout mismatch\nwant: %q\ngot:  %q", wantOut, resp.Stdout)
	}
	doc, readErr := readSessionsJSON(t, req.HomeDir)
	if readErr != nil {
		t.Fatalf("read sessions.json: %v", readErr)
	}
	var entry *sessionMapEntry
	for i := range doc.Entries {
		if doc.Entries[i].SessionID == sessionUpdateFixtureID {
			entry = &doc.Entries[i]
			break
		}
	}
	if entry == nil {
		t.Fatalf("sessions.json missing %q: %+v", sessionUpdateFixtureID, doc.Entries)
	}
	if entry.Dir != abs {
		t.Fatalf("stored dir = %q, want abs %q", entry.Dir, abs)
	}
	if !filepath.IsAbs(entry.Dir) {
		t.Fatalf("stored dir must be absolute, got %q", entry.Dir)
	}
	if entry.ChannelID != slackTestChannelID {
		t.Fatalf("channel_id not preserved: %q", entry.ChannelID)
	}
	if entry.ConfigPath != "/tmp/slack-update-cfg.json" {
		t.Fatalf("config_path not preserved: %q", entry.ConfigPath)
	}
	if entry.ThreadTS != "1710000700.000100" {
		t.Fatalf("thread_ts not preserved: %q", entry.ThreadTS)
	}
	if entry.LastMessagePreview != "before update" {
		t.Fatalf("last_message_preview not preserved: %q", entry.LastMessagePreview)
	}
	if entry.CreatedAt != "2026-07-09T10:00:00Z" {
		t.Fatalf("created_at not preserved: %q", entry.CreatedAt)
	}
	if entry.UpdatedAt == "" || entry.UpdatedAt == "2026-07-10T10:00:00Z" {
		t.Fatalf("updated_at should be bumped, got %q", entry.UpdatedAt)
	}
	// No secrets in map file.
	raw, _ := os.ReadFile(expectedSessionsJSONPath(req.HomeDir))
	for _, tok := range []string{"xoxb-", "xoxp-", "xapp-"} {
		if strings.Contains(string(raw), tok) {
			t.Fatalf("sessions.json must not embed tokens")
		}
	}
}
```
