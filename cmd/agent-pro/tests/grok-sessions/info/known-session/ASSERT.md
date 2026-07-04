## Expected

- `Info` returns populated `SessionInfo` with summary fields from `summary.json`.
- File paths include session dir, `summary.json`, `updates.jsonl (missing)`, and `signals.json`.
- `FormatInfoText` includes model, agent, messages, git, sandbox, Files, and Tokens sections.
- Relative last-active time is `2h ago` against fixed `req.Now`.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if resp.Info == nil {
		t.Fatal("info is nil")
	}

	info := resp.Info
	if info.ID != req.SessionID {
		t.Fatalf("ID = %q, want %q", info.ID, req.SessionID)
	}
	if info.Title != "Refactor auth module" {
		t.Fatalf("Title = %q, want Refactor auth module", info.Title)
	}
	if info.NumMessages != 90 || info.NumChatMessages != 40 {
		t.Fatalf("messages = %d total / %d chat, want 90 / 40", info.NumMessages, info.NumChatMessages)
	}
	if info.CurrentModelID != "grok-composer-2.5-fast" || info.AgentName != "cursor" {
		t.Fatalf("model/agent = %q / %q", info.CurrentModelID, info.AgentName)
	}
	if info.SandboxProfile != "off" {
		t.Fatalf("SandboxProfile = %q, want off", info.SandboxProfile)
	}
	if info.HeadBranch != "master-2026-07-03-1" || info.HeadCommit != "97433b50" {
		t.Fatalf("git = %s @ %s", info.HeadBranch, info.HeadCommit)
	}
	if info.SignalsPath == "" {
		t.Fatal("SignalsPath is empty")
	}
	if info.UpdatesPath == "" {
		t.Fatal("UpdatesPath is empty")
	}

	assertContains(t, resp.Output, "Session: "+req.SessionID)
	assertContains(t, resp.Output, "Title: Refactor auth module")
	assertContains(t, resp.Output, "2h ago")
	assertContains(t, resp.Output, "Model: grok-composer-2.5-fast")
	assertContains(t, resp.Output, "Agent: cursor")
	assertContains(t, resp.Output, "Messages: 90 total, 40 chat")
	assertContains(t, resp.Output, "Git: master-2026-07-03-1 @ 97433b50")
	assertContains(t, resp.Output, "Sandbox: off")
	assertContains(t, resp.Output, "Files:")
	assertContains(t, resp.Output, "summary.json")
	assertContains(t, resp.Output, "updates.jsonl (missing)")
	assertContains(t, resp.Output, "signals.json")
	assertContains(t, resp.Output, "Tokens:")
	assertContains(t, resp.Output, "Context: 75085 / 200000 (38%)")
	assertContains(t, resp.Output, "Before compaction: 0")
}
```