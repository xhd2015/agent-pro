## Preconditions
- **Reproduces**: `doctest agent implement --session-id <grok-session> --trace` shows only
  thinking blocks and ASSISTANT messages; Read/Grep/Write/Edit tool calls are missing.
- Root cause: grok tool lines never become `tool_call` entries in `events.jsonl`.
- This leaf simulates the post-fix events.jsonl that trace should display.

## Steps
1. Set up a session dir with events.jsonl containing think, tool_call, and message AgentEvents.
2. Tool calls mirror a typical implementer turn: Read requirement, Grep tree, Write file.
3. Call `traceSession`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := t.TempDir()
	sessDir := filepath.Join(dir, "sess_test")

	eventsContent := `{"type":"think","text":"I'll read the requirement, test tree, and current codebase."}
{"type":"tool_call","tool":"read","tool_input":{"path":"REQUIREMENT.md"},"text":"REQUIREMENT.md"}
{"type":"tool_call","tool":"grep","tool_input":{"pattern":"DOCTEST.md"},"text":"DOCTEST.md"}
{"type":"tool_call","tool":"write","tool_input":{"path":"libdoc/core/version.go"},"text":"libdoc/core/version.go"}
{"type":"message","text":"I'll implement the DOCTEST.md version and layout changes."}
`

	req.PreCreateDirs = []string{sessDir}
	req.PreCreateMeta = map[string]string{
		sessDir: `{"explicit_session_id":"test-grok-tool-trace","agent_runner":"grok","created_at":"2026-06-22T22:36:10Z"}`,
	}
	req.PreCreateEvents = map[string]string{
		sessDir: eventsContent,
	}
	req.SessionID = "test-grok-tool-trace"
	req.SessionBase = dir
	return nil
}
```